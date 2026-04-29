import { useEffect, useState } from "react";

import type { CreateLinkInput, Link, UpdateLinkInput } from "../types";

type Props = {
  mode: "create" | "edit";
  link: Link | null;
  isSubmitting: boolean;
  error: string;
  onClose: () => void;
  onSubmit: (payload: CreateLinkInput | UpdateLinkInput) => Promise<void>;
};

type FormState = {
  code: string;
  targetUrl: string;
  remark: string;
  tags: string;
  enabled: boolean;
};

function createInitialState(link: Link | null): FormState {
  return {
    code: link?.code ?? "",
    targetUrl: link?.target_url ?? "",
    remark: link?.remark ?? "",
    tags: link?.tags.join(", ") ?? "",
    enabled: link?.enabled ?? true,
  };
}

export function LinkFormModal({ mode, link, isSubmitting, error, onClose, onSubmit }: Props) {
  const [form, setForm] = useState<FormState>(() => createInitialState(link));
  const [validation, setValidation] = useState("");

  useEffect(() => {
    setForm(createInitialState(link));
    setValidation("");
  }, [link]);

  function updateField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!form.targetUrl.trim()) {
      setValidation("目标 URL 必填");
      return;
    }
    if (mode === "edit" && !form.code.trim()) {
      setValidation("编辑时短码不能为空");
      return;
    }

    setValidation("");

    const basePayload = {
      code: form.code.trim(),
      target_url: form.targetUrl.trim(),
      remark: form.remark.trim(),
      tags: form.tags
        .split(/,|，/)
        .map((tag) => tag.trim())
        .filter(Boolean),
    };

    if (mode === "create") {
      await onSubmit(basePayload);
      return;
    }

    await onSubmit({
      ...basePayload,
      enabled: form.enabled,
    });
  }

  return (
    <div className="modal-backdrop" role="presentation">
      <div className="modal-card" role="dialog" aria-modal="true">
        <div className="modal-header">
          <div>
            <p className="eyebrow">{mode === "create" ? "新建短链" : "编辑短链"}</p>
            <h2>{mode === "create" ? "创建跳转规则" : `编辑 ${link?.code}`}</h2>
          </div>
          <button type="button" className="modal-close-button" onClick={onClose} aria-label="关闭弹窗">
            ×
          </button>
        </div>

        <form className="modal-form" onSubmit={handleSubmit}>
          <div className="form-grid">
            <label>
              <span>短码</span>
              <input
                value={form.code}
                onChange={(event) => updateField("code", event.target.value)}
                placeholder={mode === "create" ? "留空则自动生成" : "请输入短码"}
              />
            </label>
            <label>
              <span>目标 URL</span>
              <input
                value={form.targetUrl}
                onChange={(event) => updateField("targetUrl", event.target.value)}
                placeholder="https://example.com/article"
              />
            </label>
            <label>
              <span>标签</span>
              <input
                value={form.tags}
                onChange={(event) => updateField("tags", event.target.value)}
                placeholder="多个标签用逗号分隔"
              />
            </label>
          </div>

          <label>
            <span>备注</span>
            <textarea
              value={form.remark}
              onChange={(event) => updateField("remark", event.target.value)}
              rows={4}
              placeholder="给这条短链补充用途、投放位置或备注说明"
            />
          </label>

          {mode === "edit" && (
            <label className="inline-checkbox">
              <input
                type="checkbox"
                checked={form.enabled}
                onChange={(event) => updateField("enabled", event.target.checked)}
              />
              <span>启用该短链</span>
            </label>
          )}

          {(validation || error) && <p className="error-banner">{validation || error}</p>}

          <div className="modal-actions">
            <button type="submit" className="primary-button" disabled={isSubmitting}>
              {isSubmitting ? "提交中..." : mode === "create" ? "创建短链" : "保存修改"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
