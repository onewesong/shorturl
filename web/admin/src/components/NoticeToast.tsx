type NoticeTone = "success" | "error";

export type Notice = {
  id: number;
  tone: NoticeTone;
  title: string;
  message: string;
};

type Props = {
  notice: Notice | null;
  onDismiss: () => void;
};

export function NoticeToast({ notice, onDismiss }: Props) {
  if (!notice) {
    return null;
  }

  return (
    <div className="notice-region" aria-live={notice.tone === "error" ? "assertive" : "polite"}>
      <section className={`notice-toast notice-toast-${notice.tone}`} role={notice.tone === "error" ? "alert" : "status"}>
        <span className="notice-icon" aria-hidden="true">
          {notice.tone === "success" ? <CheckIcon /> : <AlertIcon />}
        </span>
        <div className="notice-content">
          <strong>{notice.title}</strong>
          <p>{notice.message}</p>
        </div>
        <button type="button" className="notice-close-button" onClick={onDismiss} aria-label="关闭通知">
          ×
        </button>
      </section>
    </div>
  );
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 24 24">
      <path d="m6.2 12.4 3.8 3.8 7.8-8.4" />
    </svg>
  );
}

function AlertIcon() {
  return (
    <svg viewBox="0 0 24 24">
      <path d="M12 8v5" />
      <path d="M12 17h.01" />
      <path d="M10.2 4.5 3.2 17a2 2 0 0 0 1.8 3h14a2 2 0 0 0 1.8-3l-7-12.5a2 2 0 0 0-3.6 0Z" />
    </svg>
  );
}
