import { useState } from "react";

type Props = {
  isSubmitting: boolean;
  error: string;
  onSubmit: (username: string, password: string) => Promise<void>;
};

export function LoginPage({ isSubmitting, error, onSubmit }: Props) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [validation, setValidation] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!username.trim()) {
      setValidation("请输入管理员用户名");
      return;
    }
    if (!password) {
      setValidation("请输入管理员密码");
      return;
    }

    setValidation("");
    await onSubmit(username.trim(), password);
  }

  return (
    <main className="login-shell">
      <section className="login-card">
        <p className="eyebrow">shorturl Admin</p>
        <h1>短链后台登录</h1>
        <p className="muted-copy">使用部署时初始化的管理员账号密码登录，进入短链管理台。</p>

        <form onSubmit={handleSubmit} className="login-form">
          <label>
            <span>用户名</span>
            <input
              type="text"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="请输入管理员用户名"
            />
          </label>

          <label>
            <span>密码</span>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="请输入管理员密码"
            />
          </label>

          {(validation || error) && <p className="error-banner">{validation || error}</p>}

          <button type="submit" className="primary-button wide-button" disabled={isSubmitting}>
            {isSubmitting ? "登录中..." : "进入后台"}
          </button>
        </form>
      </section>
    </main>
  );
}
