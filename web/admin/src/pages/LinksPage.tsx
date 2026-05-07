import { useEffect, useRef, useState } from "react";

import { ApiError, createLink, deleteLink, getLinkAnalytics, updateLink } from "../lib/api";
import type { AuthSession, CreateLinkInput, Link, LinkAnalytics, UpdateLinkInput } from "../types";
import { LinkFormModal } from "../components/LinkFormModal";
import { NoticeToast, type Notice } from "../components/NoticeToast";

type Props = {
  session: AuthSession;
  links: Link[];
  isLoading: boolean;
  error: string;
  onReload: () => Promise<void>;
  onLogout: () => void;
};

type ModalState = { type: "create" } | { type: "edit"; link: Link } | null;
type SortKey = "code" | "target_url" | "tags" | "enabled" | "click_count";
type SortDirection = "asc" | "desc";

export function LinksPage({ session, links, isLoading, error, onReload, onLogout }: Props) {
  const [modal, setModal] = useState<ModalState>(null);
  const [activeTag, setActiveTag] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [sortState, setSortState] = useState<{ key: SortKey; direction: SortDirection } | null>(null);
  const [actionError, setActionError] = useState("");
  const [notice, setNotice] = useState<Notice | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [analyticsDays, setAnalyticsDays] = useState(7);
  const [selectedLinkId, setSelectedLinkId] = useState<number | null>(null);
  const [analytics, setAnalytics] = useState<LinkAnalytics | null>(null);
  const [analyticsError, setAnalyticsError] = useState("");
  const [isLoadingAnalytics, setIsLoadingAnalytics] = useState(false);
  const [copiedLinkId, setCopiedLinkId] = useState<number | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const copyResetTimerRef = useRef<number | null>(null);
  const noticeTimerRef = useRef<number | null>(null);

  const availableTags = Array.from(new Set(links.flatMap((link) => link.tags))).sort((left, right) => left.localeCompare(right));
  const searchTokens = parseSearchQuery(searchQuery);
  const normalizedSearchQuery = searchQuery.trim().toLowerCase();
  const filteredLinks = links.filter((link) => {
    const matchesTag = activeTag ? link.tags.includes(activeTag) : true;
    const searchableText = [link.code, link.target_url, link.remark, ...link.tags].join(" ").toLowerCase();
    const matchesText = searchTokens.terms.every((term) => searchableText.includes(term));
    const matchesSearchTags = searchTokens.tags.every((tag) => link.tags.some((linkTag) => linkTag.toLowerCase().includes(tag)));
    const matchesSearch = matchesText && matchesSearchTags;

    return matchesTag && matchesSearch;
  });
  const sortedLinks = sortState ? [...filteredLinks].sort((left, right) => compareLinks(left, right, sortState)) : filteredLinks;
  const selectedLink = sortedLinks.find((item) => item.id === selectedLinkId) ?? links.find((item) => item.id === selectedLinkId) ?? null;

  const summary = {
    total: filteredLinks.length,
    enabled: filteredLinks.filter((link) => link.enabled).length,
    clicks: filteredLinks.reduce((total, link) => total + link.click_count, 0),
  };
  const summaryBars = createSummaryBars(filteredLinks);
  const enabledRate = summary.total > 0 ? Math.round((summary.enabled / summary.total) * 100) : 0;
  const averageClicks = summary.total > 0 ? Math.round(summary.clicks / summary.total) : 0;
  const emptyStateMessage = normalizedSearchQuery
    ? `没有匹配关键词 ${searchQuery.trim()} 的短链`
    : activeTag
      ? `没有命中标签 ${activeTag} 的短链`
      : "暂无短链";

  useEffect(() => {
    searchInputRef.current?.focus();
  }, []);

  useEffect(() => {
    function handleSearchShortcut(event: KeyboardEvent) {
      if (event.key !== "/" || modal || selectedLinkId) {
        return;
      }

      const target = event.target;
      if (target instanceof HTMLElement) {
        const tagName = target.tagName.toLowerCase();
        if (target.isContentEditable || tagName === "input" || tagName === "textarea" || tagName === "select") {
          return;
        }
      }

      event.preventDefault();
      searchInputRef.current?.focus();
      searchInputRef.current?.select();
    }

    window.addEventListener("keydown", handleSearchShortcut);
    return () => window.removeEventListener("keydown", handleSearchShortcut);
  }, [modal, selectedLinkId]);

  useEffect(() => {
    return () => {
      if (copyResetTimerRef.current) {
        window.clearTimeout(copyResetTimerRef.current);
      }
      if (noticeTimerRef.current) {
        window.clearTimeout(noticeTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (error) {
      showNotice("error", "列表加载失败", error);
    }
  }, [error]);

  useEffect(() => {
    if (!selectedLinkId) {
      setAnalytics(null);
      setAnalyticsError("");
      return;
    }

    const linkId = selectedLinkId;
    let cancelled = false;

    async function loadAnalytics() {
      setIsLoadingAnalytics(true);
      setAnalyticsError("");

      try {
        const result = await getLinkAnalytics(linkId, analyticsDays);
        if (cancelled) {
          return;
        }
        setAnalytics(result);
      } catch (unknownError) {
        if (cancelled) {
          return;
        }

        if (unknownError instanceof ApiError) {
          setAnalyticsError(unknownError.message);
        } else if (unknownError instanceof Error) {
          setAnalyticsError(unknownError.message);
        } else {
          setAnalyticsError("unknown_error");
        }
        setAnalytics(null);
      } finally {
        if (!cancelled) {
          setIsLoadingAnalytics(false);
        }
      }
    }

    void loadAnalytics();

    return () => {
      cancelled = true;
    };
  }, [analyticsDays, selectedLinkId]);

  function showNotice(tone: Notice["tone"], title: string, message: string) {
    const id = Date.now();

    if (noticeTimerRef.current) {
      window.clearTimeout(noticeTimerRef.current);
    }

    setNotice({ id, tone, title, message });
    noticeTimerRef.current = window.setTimeout(
      () => {
        setNotice((current) => (current?.id === id ? null : current));
      },
      tone === "success" ? 3600 : 6200,
    );
  }

  function dismissNotice() {
    if (noticeTimerRef.current) {
      window.clearTimeout(noticeTimerRef.current);
      noticeTimerRef.current = null;
    }

    setNotice(null);
  }

  async function runAction(action: () => Promise<void>, successText: string, errorTarget: "notice" | "form" = "notice") {
    setIsSubmitting(true);
    setActionError("");

    try {
      await action();
      showNotice("success", successText, "列表已同步到最新状态。");
      await onReload();
    } catch (unknownError) {
      const message = getActionErrorMessage(unknownError);
      setActionError(message);
      if (errorTarget === "notice") {
        showNotice("error", "操作失败", message);
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleCreate(payload: CreateLinkInput) {
    await runAction(async () => {
      await createLink(payload);
      setModal(null);
    }, "短链已创建", "form");
  }

  async function handleUpdate(linkId: number, payload: UpdateLinkInput) {
    await runAction(async () => {
      await updateLink(linkId, payload);
      setModal(null);
    }, "短链已更新", "form");
  }

  function handleSort(key: SortKey) {
    setSortState((current) => {
      if (current?.key !== key) {
        return { key, direction: "asc" };
      }

      return { key, direction: current.direction === "asc" ? "desc" : "asc" };
    });
  }

  async function handleToggleEnabled(link: Link) {
    await runAction(async () => {
      await updateLink(link.id, {
        code: link.code,
        target_url: link.target_url,
        remark: link.remark,
        tags: link.tags,
        enabled: !link.enabled,
      });
    }, link.enabled ? "短链已禁用" : "短链已启用");
  }

  async function handleDelete(link: Link) {
    if (!window.confirm(`确认删除短链 ${link.code}？此操作不可恢复。`)) {
      return;
    }

    await runAction(async () => {
      await deleteLink(link.id);
      if (selectedLinkId === link.id) {
        setSelectedLinkId(null);
        setAnalytics(null);
        setAnalyticsError("");
      }
    }, "短链已删除");
  }

  async function handleCopyShortLink(link: Link) {
    setActionError("");

    if (!navigator.clipboard) {
      showNotice("error", "复制失败", "当前浏览器不支持一键复制。");
      return;
    }

    const shortUrl = new URL(`/${link.code}`, window.location.origin).toString();

    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopiedLinkId(link.id);

      if (copyResetTimerRef.current) {
        window.clearTimeout(copyResetTimerRef.current);
      }

      copyResetTimerRef.current = window.setTimeout(() => {
        setCopiedLinkId(null);
      }, 1600);
      showNotice("success", "已复制短链", shortUrl);
    } catch {
      showNotice("error", "复制失败", "复制短链失败，请手动复制短码。");
    }
  }

  return (
    <main className="dashboard-shell">
      <NoticeToast notice={notice} onDismiss={dismissNotice} />
      <section className="hero-panel">
        <div className="hero-brand">
          <h1 className="admin-logo-lockup" aria-label="ShortURL Admin 短链管理后台">
            <span className="admin-logo-mark" aria-hidden="true">
              <svg viewBox="0 0 48 48">
                <path d="M17.2 30.8h-3.4a8.8 8.8 0 0 1 0-17.6h7.8" />
                <path d="M30.8 17.2h3.4a8.8 8.8 0 1 1 0 17.6h-7.8" />
                <path d="M18.5 24h11" />
                <path d="M9.8 38.2 38.2 9.8" />
              </svg>
            </span>
            <span className="admin-logo-copy">
              <span className="admin-logo-name">ShortURL</span>
              <span className="admin-logo-subtitle">短链管理</span>
            </span>
          </h1>
        </div>

        <div className="hero-actions">
          <span className="header-account-chip" title={`当前登录账号：${session.username}`}>
            <span aria-hidden="true">{session.username.slice(0, 1).toUpperCase()}</span>
            <strong>{session.username}</strong>
          </span>
          <button type="button" className="ghost-button header-icon-button" onClick={onLogout} aria-label="退出登录" title="退出登录">
            <LogoutIcon />
          </button>
        </div>
      </section>

      <section className="summary-grid">
        <article className="summary-card">
          <div>
            <span className="summary-label">短链总数</span>
            <strong className="numeric-value">{summary.total}</strong>
          </div>
          <MiniSparkline bars={summaryBars.total} />
          <span className="summary-note">当前视图 · {filteredLinks.length} 条</span>
        </article>
        <article className="summary-card">
          <div>
            <span className="summary-label">启用中</span>
            <strong className="numeric-value">{summary.enabled}</strong>
          </div>
          <MiniSparkline bars={summaryBars.enabled} tone="green" />
          <span className="summary-note">启用率 {enabledRate}%</span>
        </article>
        <article className="summary-card">
          <div>
            <span className="summary-label">累计点击</span>
            <strong className="numeric-value">{summary.clicks}</strong>
          </div>
          <MiniSparkline bars={summaryBars.clicks} tone="violet" />
          <span className="summary-note">平均 {averageClicks} 次 / 链接</span>
        </article>
      </section>

      <section className="table-card">
        <div className="table-header links-table-header">
          <div>
            <p className="eyebrow">Links</p>
            <h2>跳转规则列表</h2>
            {(activeTag || normalizedSearchQuery) && (
              <p className="muted-copy">
                当前按{activeTag && <>标签 <strong>{activeTag}</strong></>}
                {activeTag && normalizedSearchQuery && " 和"}
                {normalizedSearchQuery && <>关键词 <strong>{searchQuery.trim()}</strong></>}筛选
              </p>
            )}
          </div>
          <label className="search-control table-search-control">
            <span className="visually-hidden">搜索短链</span>
            <input
              ref={searchInputRef}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              aria-label="搜索短链"
              placeholder="搜索短码、目标地址、备注或标签"
              autoFocus
            />
            <span className="search-shortcut" aria-hidden="true">/</span>
            {searchQuery && (
              <button type="button" className="search-clear-button" onClick={() => setSearchQuery("")} aria-label="清空搜索">
                ×
              </button>
            )}
          </label>
          <div className="table-header-actions">
            <button type="button" className="ghost-button header-icon-button" onClick={onReload} disabled={isLoading} aria-label="刷新列表" title="刷新列表">
              <RefreshIcon />
            </button>
            <button type="button" className="primary-button table-create-button" onClick={() => setModal({ type: "create" })}>
              <PlusIcon />
              <span>新建</span>
            </button>
          </div>
        </div>

        {availableTags.length > 0 && (
          <div className="filter-bar">
            <button
              type="button"
              className={`tag-filter ${activeTag === "" ? "tag-filter-active" : ""}`}
              onClick={() => setActiveTag("")}
            >
              全部
            </button>
            {availableTags.map((tag) => (
              <button
                key={tag}
                type="button"
                className={`tag-filter ${activeTag === tag ? "tag-filter-active" : ""}`}
                onClick={() => setActiveTag(tag)}
              >
                #{tag}
              </button>
            ))}
          </div>
        )}

        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th scope="col" aria-sort={getAriaSort(sortState, "code")}>
                  <SortHeader label="短码" sortKey="code" sortState={sortState} onSort={handleSort} />
                </th>
                <th scope="col" aria-sort={getAriaSort(sortState, "target_url")}>
                  <SortHeader label="目标地址" sortKey="target_url" sortState={sortState} onSort={handleSort} />
                </th>
                <th scope="col" aria-sort={getAriaSort(sortState, "tags")}>
                  <SortHeader label="标签" sortKey="tags" sortState={sortState} onSort={handleSort} />
                </th>
                <th scope="col" aria-sort={getAriaSort(sortState, "enabled")}>
                  <SortHeader label="状态" sortKey="enabled" sortState={sortState} onSort={handleSort} />
                </th>
                <th scope="col" aria-sort={getAriaSort(sortState, "click_count")}>
                  <SortHeader label="点击" sortKey="click_count" sortState={sortState} onSort={handleSort} />
                </th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {sortedLinks.map((link) => (
                <tr key={link.id}>
                  <td>
                    <div className="account-title">
                      <a
                        className="shortcode-link"
                        href={`/${link.code}`}
                        target="_blank"
                        rel="noreferrer"
                        title={`打开 /${link.code}`}
                      >
                        {link.code}
                      </a>
                      <button
                        type="button"
                        className="copy-link-button"
                        onClick={() => void handleCopyShortLink(link)}
                        aria-label={`复制 ${link.code} 短链`}
                        title={copiedLinkId === link.id ? "已复制" : "复制短链"}
                      >
                        {copiedLinkId === link.id ? (
                          <span aria-hidden="true">✓</span>
                        ) : (
                          <svg viewBox="0 0 24 24" aria-hidden="true">
                            <path d="M8 8.5a2.5 2.5 0 0 1 2.5-2.5h6A2.5 2.5 0 0 1 19 8.5v6a2.5 2.5 0 0 1-2.5 2.5h-6A2.5 2.5 0 0 1 8 14.5v-6Z" />
                            <path d="M6 14.5h-.5A2.5 2.5 0 0 1 3 12V5.5A2.5 2.5 0 0 1 5.5 3H12a2.5 2.5 0 0 1 2.5 2.5V6" />
                          </svg>
                        )}
                      </button>
                    </div>
                  </td>
                  <td className="target-url-cell">
                    <div className="stacked-meta">
                      <span className="target-url-text" title={link.target_url}>
                        {link.target_url}
                      </span>
                      {link.remark && (
                        <span className="remark-text" title={link.remark}>
                          {link.remark}
                        </span>
                      )}
                    </div>
                  </td>
                  <td>
                    <div className="tag-list">
                      {link.tags.length > 0 ? (
                        link.tags.map((tag) => (
                          <button
                            key={tag}
                            type="button"
                            className={`inline-tag ${activeTag === tag ? "inline-tag-active" : ""}`}
                            onClick={() => setActiveTag(tag)}
                            title={`按标签 ${tag} 筛选`}
                          >
                            #{tag}
                          </button>
                        ))
                      ) : (
                        <span className="tag-empty">-</span>
                      )}
                    </div>
                  </td>
                  <td>
                    <div className="badge-row">
                      <button
                        type="button"
                        className={`switch-button ${link.enabled ? "switch-button-on" : ""}`}
                        onClick={() => void handleToggleEnabled(link)}
                        disabled={isSubmitting}
                        aria-pressed={link.enabled}
                        aria-label={`${link.enabled ? "禁用" : "启用"} ${link.code}`}
                        title={link.enabled ? "点击禁用" : "点击启用"}
                      >
                        <span aria-hidden="true" />
                        <strong>{link.enabled ? "启用" : "禁用"}</strong>
                      </button>
                    </div>
                  </td>
                  <td>
                    <div className="stacked-meta numeric-value">
                      <span>{link.click_count}</span>
                    </div>
                  </td>
                  <td>
                    <div className="table-actions">
                      <button
                        type="button"
                        className="icon-action-button"
                        onClick={() => setSelectedLinkId(link.id)}
                        aria-label="分析"
                        title="分析"
                      >
                        <ChartIcon />
                      </button>
                      <button
                        type="button"
                        className="icon-action-button"
                        onClick={() => setModal({ type: "edit", link })}
                        aria-label="编辑"
                        title="编辑"
                      >
                        <EditIcon />
                      </button>
                      <button
                        type="button"
                        className="icon-action-button icon-action-button-danger"
                        onClick={() => void handleDelete(link)}
                        disabled={isSubmitting}
                        aria-label={`删除 ${link.code}`}
                        title="删除"
                      >
                        <TrashIcon />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {filteredLinks.length === 0 && (
                <tr>
                  <td colSpan={6}>
                    <div className="empty-state">{emptyStateMessage}</div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      {selectedLink && (
        <div className="modal-backdrop" role="presentation">
          <section className="modal-card analytics-modal-card" role="dialog" aria-modal="true" aria-labelledby="analytics-modal-title">
            <div className="modal-header">
              <div>
                <p className="eyebrow">Analytics</p>
                <h2 id="analytics-modal-title">{selectedLink.code} 的访问分析</h2>
                <p className="muted-copy">
                  看最近访问 IP、来源域名、客户端类型，以及 {analyticsDays} 天访问曲线。
                </p>
              </div>
              <button
                type="button"
                className="modal-close-button"
                onClick={() => {
                  setSelectedLinkId(null);
                  setAnalytics(null);
                  setAnalyticsError("");
                }}
                aria-label="关闭分析弹窗"
              >
                ×
              </button>
            </div>

            <div className="analytics-modal-toolbar">
              {[7, 30].map((days) => (
                <button
                  key={days}
                  type="button"
                  className={`tag-filter ${analyticsDays === days ? "tag-filter-active" : ""}`}
                  onClick={() => setAnalyticsDays(days)}
                >
                  近 {days} 天
                </button>
              ))}
            </div>

            {analyticsError && <p className="error-banner">{analyticsError}</p>}
            {isLoadingAnalytics && <p className="muted-copy">正在加载分析数据...</p>}

            {analytics && (
              <>
                <section className="analytics-summary-grid">
                  <article className="summary-card">
                    <span className="summary-label">窗口点击</span>
                    <strong>{analytics.recent_clicks}</strong>
                  </article>
                  <article className="summary-card">
                    <span className="summary-label">独立 IP</span>
                    <strong>{analytics.unique_ips}</strong>
                  </article>
                  <article className="summary-card">
                    <span className="summary-label">累计点击</span>
                    <strong>{analytics.link.click_count}</strong>
                  </article>
                  <article className="summary-card">
                    <span className="summary-label">最近访问</span>
                    <strong className="summary-meta">
                      {analytics.last_visited_at ? formatDateTime(analytics.last_visited_at) : "暂无"}
                    </strong>
                  </article>
                </section>

                <section className="analytics-grid">
                  <article className="insight-card">
                    <div className="insight-header">
                      <h3>访问曲线</h3>
                      <span className="status-pill">{analytics.range_days} 天</span>
                    </div>
                    <div className="sparkline" aria-label="访问曲线">
                      {analytics.time_series.map((point) => (
                        <div key={point.bucket} className="sparkline-item">
                          <div
                            className="sparkline-bar"
                            style={{ height: `${Math.max((point.clicks / maxClicks(analytics.time_series)) * 100, point.clicks > 0 ? 16 : 4)}%` }}
                            title={`${point.bucket} · ${point.clicks} 次`}
                          />
                          <span>{point.bucket.slice(5)}</span>
                        </div>
                      ))}
                    </div>
                  </article>

                  <article className="insight-card">
                    <div className="insight-header">
                      <h3>来源分布</h3>
                      <span className="muted-copy">Top Referrers</span>
                    </div>
                    <div className="breakdown-list">
                      {analytics.top_referrers.length > 0 ? (
                        analytics.top_referrers.map((item) => (
                          <div key={item.name} className="breakdown-row">
                            <span>{item.name}</span>
                            <strong>{item.count}</strong>
                          </div>
                        ))
                      ) : (
                        <p className="muted-copy">暂无来源数据</p>
                      )}
                    </div>
                  </article>

                  <article className="insight-card">
                    <div className="insight-header">
                      <h3>客户端分布</h3>
                      <span className="muted-copy">Top Clients</span>
                    </div>
                    <div className="breakdown-list">
                      {analytics.top_clients.length > 0 ? (
                        analytics.top_clients.map((item) => (
                          <div key={item.name} className="breakdown-row">
                            <span>{item.name}</span>
                            <strong>{item.count}</strong>
                          </div>
                        ))
                      ) : (
                        <p className="muted-copy">暂无客户端数据</p>
                      )}
                    </div>
                  </article>
                </section>

                <section className="insight-card">
                  <div className="insight-header">
                    <h3>最近访问明细</h3>
                    <span className="muted-copy">按时间倒序展示最近 20 次访问</span>
                  </div>
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr>
                          <th>访问时间</th>
                          <th>来源 IP</th>
                          <th>来源域名</th>
                          <th>客户端</th>
                          <th>设备 / 系统</th>
                        </tr>
                      </thead>
                      <tbody>
                        {analytics.recent_visits.map((visit) => (
                          <tr key={`${visit.visited_at}-${visit.ip_masked}-${visit.user_agent}`}>
                            <td>{formatDateTime(visit.visited_at)}</td>
                            <td>{visit.ip_masked || "未知"}</td>
                            <td title={visit.referer || "直接访问"}>{visit.referer_host || "直接访问"}</td>
                            <td>{visit.client_name}</td>
                            <td>
                              {visit.device_type} / {visit.os}
                            </td>
                          </tr>
                        ))}
                        {analytics.recent_visits.length === 0 && (
                          <tr>
                            <td colSpan={5}>
                              <div className="empty-state">该短链还没有访问明细</div>
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </section>
              </>
            )}
          </section>
        </div>
      )}

      {modal?.type === "create" && (
        <LinkFormModal
          mode="create"
          link={null}
          isSubmitting={isSubmitting}
          error={actionError}
          onClose={() => {
            setModal(null);
            setActionError("");
          }}
          onSubmit={(payload) => handleCreate(payload as CreateLinkInput)}
        />
      )}

      {modal?.type === "edit" && (
        <LinkFormModal
          mode="edit"
          link={modal.link}
          isSubmitting={isSubmitting}
          error={actionError}
          onClose={() => {
            setModal(null);
            setActionError("");
          }}
          onSubmit={(payload) => handleUpdate(modal.link.id, payload as UpdateLinkInput)}
        />
      )}
    </main>
  );
}

function maxClicks(points: LinkAnalytics["time_series"]) {
  return points.reduce((value, point) => Math.max(value, point.clicks), 1);
}

function getActionErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }

  return "unknown_error";
}

function parseSearchQuery(value: string) {
  return value
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .reduce(
      (tokens, item) => {
        if (item.startsWith("tag:") && item.length > 4) {
          tokens.tags.push(item.slice(4));
        } else {
          tokens.terms.push(item);
        }

        return tokens;
      },
      { terms: [] as string[], tags: [] as string[] },
    );
}

function createSummaryBars(items: Link[]) {
  const base = items.length > 0 ? items.slice(0, 8) : [{ click_count: 0, enabled: false } as Link];
  const maxClickCount = base.reduce((value, link) => Math.max(value, link.click_count), 1);

  return {
    total: base.map((_, index) => 28 + ((index + 1) % 4) * 14),
    enabled: base.map((link, index) => (link.enabled ? 44 + (index % 3) * 12 : 18)),
    clicks: base.map((link) => Math.max((link.click_count / maxClickCount) * 100, link.click_count > 0 ? 22 : 8)),
  };
}

function compareLinks(left: Link, right: Link, sortState: { key: SortKey; direction: SortDirection }) {
  const directionFactor = sortState.direction === "asc" ? 1 : -1;
  let result = 0;

  switch (sortState.key) {
    case "code":
      result = left.code.localeCompare(right.code, "zh-CN", { numeric: true, sensitivity: "base" });
      break;
    case "target_url":
      result = left.target_url.localeCompare(right.target_url, "zh-CN", { numeric: true, sensitivity: "base" });
      break;
    case "tags":
      result = left.tags.join(" ").localeCompare(right.tags.join(" "), "zh-CN", { numeric: true, sensitivity: "base" });
      break;
    case "enabled":
      result = Number(left.enabled) - Number(right.enabled);
      break;
    case "click_count":
      result = left.click_count - right.click_count;
      break;
  }

  if (result === 0) {
    result = left.code.localeCompare(right.code, "zh-CN", { numeric: true, sensitivity: "base" });
  }

  return result * directionFactor;
}

function getAriaSort(sortState: { key: SortKey; direction: SortDirection } | null, key: SortKey) {
  if (sortState?.key !== key) {
    return "none";
  }

  return sortState.direction === "asc" ? "ascending" : "descending";
}

function SortHeader({
  label,
  sortKey,
  sortState,
  onSort,
}: {
  label: string;
  sortKey: SortKey;
  sortState: { key: SortKey; direction: SortDirection } | null;
  onSort: (key: SortKey) => void;
}) {
  const isActive = sortState?.key === sortKey;

  return (
    <button type="button" className={`sort-header-button ${isActive ? "sort-header-button-active" : ""}`} onClick={() => onSort(sortKey)}>
      <span>{label}</span>
      <span className="sort-indicator" aria-hidden="true">
        {isActive ? (sortState.direction === "asc" ? "↑" : "↓") : "↕"}
      </span>
    </button>
  );
}

function MiniSparkline({ bars, tone = "blue" }: { bars: number[]; tone?: "blue" | "green" | "violet" }) {
  return (
    <div className={`summary-sparkline summary-sparkline-${tone}`} aria-hidden="true">
      {bars.map((height, index) => (
        <span key={index} style={{ height: `${height}%` }} />
      ))}
    </div>
  );
}

function RefreshIcon() {
  return (
    <svg className="button-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M20 11a8 8 0 0 0-14.2-5" />
      <path d="M4 5v5h5" />
      <path d="M4 13a8 8 0 0 0 14.2 5" />
      <path d="M20 19v-5h-5" />
    </svg>
  );
}

function PlusIcon() {
  return (
    <svg className="button-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 5v14" />
      <path d="M5 12h14" />
    </svg>
  );
}

function LogoutIcon() {
  return (
    <svg className="button-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M10 6H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h4" />
      <path d="M14 16l4-4-4-4" />
      <path d="M18 12H9" />
    </svg>
  );
}

function ChartIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4 19V5" />
      <path d="M4 19h16" />
      <path d="m7 15 3.5-4 3 2.5L19 7" />
    </svg>
  );
}

function EditIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4 20h4l10.5-10.5a2.1 2.1 0 0 0-3-3L5 17v3Z" />
      <path d="m14 8 2 2" />
    </svg>
  );
}

function TrashIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4 7h16" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
      <path d="M6 7l1 13h10l1-13" />
      <path d="M9 7V4h6v3" />
    </svg>
  );
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
