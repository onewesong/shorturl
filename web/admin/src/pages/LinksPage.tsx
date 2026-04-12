import { useEffect, useState } from "react";

import { ApiError, createLink, getLinkAnalytics, updateLink } from "../lib/api";
import type { AuthSession, CreateLinkInput, Link, LinkAnalytics, UpdateLinkInput } from "../types";
import { LinkFormModal } from "../components/LinkFormModal";

type Props = {
  session: AuthSession;
  links: Link[];
  isLoading: boolean;
  error: string;
  onReload: () => Promise<void>;
  onLogout: () => void;
};

type ModalState = { type: "create" } | { type: "edit"; link: Link } | null;

export function LinksPage({ session, links, isLoading, error, onReload, onLogout }: Props) {
  const [modal, setModal] = useState<ModalState>(null);
  const [activeTag, setActiveTag] = useState("");
  const [actionError, setActionError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [analyticsDays, setAnalyticsDays] = useState(7);
  const [selectedLinkId, setSelectedLinkId] = useState<number | null>(null);
  const [analytics, setAnalytics] = useState<LinkAnalytics | null>(null);
  const [analyticsError, setAnalyticsError] = useState("");
  const [isLoadingAnalytics, setIsLoadingAnalytics] = useState(false);

  const availableTags = Array.from(new Set(links.flatMap((link) => link.tags))).sort((left, right) => left.localeCompare(right));
  const filteredLinks = activeTag ? links.filter((link) => link.tags.includes(activeTag)) : links;
  const selectedLink = filteredLinks.find((item) => item.id === selectedLinkId) ?? links.find((item) => item.id === selectedLinkId) ?? null;

  const summary = {
    total: filteredLinks.length,
    enabled: filteredLinks.filter((link) => link.enabled).length,
    clicks: filteredLinks.reduce((total, link) => total + link.click_count, 0),
  };

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

  async function runAction(action: () => Promise<void>, successText: string) {
    setIsSubmitting(true);
    setActionError("");
    setSuccessMessage("");

    try {
      await action();
      setSuccessMessage(successText);
      await onReload();
    } catch (unknownError) {
      if (unknownError instanceof ApiError) {
        setActionError(unknownError.message);
      } else if (unknownError instanceof Error) {
        setActionError(unknownError.message);
      } else {
        setActionError("unknown_error");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleCreate(payload: CreateLinkInput) {
    await runAction(async () => {
      await createLink(payload);
      setModal(null);
    }, "短链已创建");
  }

  async function handleUpdate(linkId: number, payload: UpdateLinkInput) {
    await runAction(async () => {
      await updateLink(linkId, payload);
      setModal(null);
    }, "短链已更新");
  }

  return (
    <main className="dashboard-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">shorturl / admin</p>
          <h1>短链管理后台</h1>
          <p className="muted-copy">
            当前登录账号：<strong>{session.username}</strong>。在这里维护短码、目标地址、启用状态和点击数据。
          </p>
        </div>

        <div className="hero-actions">
          <button type="button" className="secondary-button" onClick={onReload} disabled={isLoading}>
            刷新列表
          </button>
          <button type="button" className="primary-button" onClick={() => setModal({ type: "create" })}>
            新建短链
          </button>
          <button type="button" className="ghost-button" onClick={onLogout}>
            退出登录
          </button>
        </div>
      </section>

      <section className="summary-grid">
        <article className="summary-card">
          <span className="summary-label">短链总数</span>
          <strong>{summary.total}</strong>
        </article>
        <article className="summary-card">
          <span className="summary-label">启用中</span>
          <strong>{summary.enabled}</strong>
        </article>
        <article className="summary-card">
          <span className="summary-label">累计点击</span>
          <strong>{summary.clicks}</strong>
        </article>
      </section>

      {(error || actionError) && <p className="error-banner">{error || actionError}</p>}
      {successMessage && <p className="success-banner">{successMessage}</p>}

      <section className="table-card">
        <div className="table-header">
          <div>
            <p className="eyebrow">Links</p>
            <h2>跳转规则列表</h2>
            {activeTag && (
              <p className="muted-copy">
                当前按标签 <strong>{activeTag}</strong> 筛选
              </p>
            )}
          </div>
          <span className="status-pill">{isLoading ? "加载中..." : `${filteredLinks.length} links`}</span>
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
                <th>短码</th>
                <th>目标地址</th>
                <th>标签</th>
                <th>状态</th>
                <th>点击</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredLinks.map((link) => (
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
                    </div>
                  </td>
                  <td className="target-url-cell">
                    <div className="stacked-meta">
                      <span className="target-url-text" title={link.target_url}>
                        {link.target_url}
                      </span>
                      <span className="remark-text" title={link.remark}>
                        {link.remark || "暂无备注"}
                      </span>
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
                        <span className="tag-empty">无标签</span>
                      )}
                    </div>
                  </td>
                  <td>
                    <div className="badge-row">
                      <span className={`badge ${link.enabled ? "badge-enabled" : "badge-disabled"}`}>
                        {link.enabled ? "启用" : "禁用"}
                      </span>
                    </div>
                  </td>
                  <td>
                    <div className="stacked-meta">
                      <span>{link.click_count}</span>
                    </div>
                  </td>
                  <td>
                    <div className="table-actions">
                      <button
                        type="button"
                        className="secondary-button small-button"
                        onClick={() => setSelectedLinkId(link.id)}
                      >
                        分析
                      </button>
                      <button
                        type="button"
                        className="ghost-button small-button"
                        onClick={() => setModal({ type: "edit", link })}
                      >
                        编辑
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {filteredLinks.length === 0 && (
                <tr>
                  <td colSpan={6}>
                    <div className="empty-state">
                      {activeTag ? `没有命中标签 ${activeTag} 的短链` : "暂无短链"}
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      {selectedLink && (
        <section className="analytics-card">
          <div className="table-header">
            <div>
              <p className="eyebrow">Analytics</p>
              <h2>{selectedLink.code} 的访问分析</h2>
              <p className="muted-copy">
                看最近访问 IP、来源域名、客户端类型，以及 {analyticsDays} 天访问曲线。
              </p>
            </div>
            <div className="hero-actions">
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

function formatDateTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
