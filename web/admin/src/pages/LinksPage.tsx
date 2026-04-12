import { useState } from "react";

import { ApiError, createLink, updateLink } from "../lib/api";
import type { AuthSession, CreateLinkInput, Link, UpdateLinkInput } from "../types";
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

  const availableTags = Array.from(new Set(links.flatMap((link) => link.tags))).sort((left, right) => left.localeCompare(right));
  const filteredLinks = activeTag ? links.filter((link) => link.tags.includes(activeTag)) : links;

  const summary = {
    total: filteredLinks.length,
    enabled: filteredLinks.filter((link) => link.enabled).length,
    clicks: filteredLinks.reduce((total, link) => total + link.click_count, 0),
  };

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
