import { useEffect, useState } from "react";
import { Navigate, Route, Routes, useNavigate } from "react-router-dom";

import { getErrorMessage, isUnauthorizedError } from "./lib/auth";
import { getSession, listLinks, login, logout } from "./lib/api";
import { LoginPage } from "./pages/LoginPage";
import { LinksPage } from "./pages/LinksPage";
import type { AuthSession, Link } from "./types";

function ProtectedRoute({ session, children }: { session: AuthSession | null; children: JSX.Element }) {
  if (!session) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

export default function App() {
  const navigate = useNavigate();
  const [session, setSession] = useState<AuthSession | null>(null);
  const [links, setLinks] = useState<Link[]>([]);
  const [isInitializing, setIsInitializing] = useState(true);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [isLoadingLinks, setIsLoadingLinks] = useState(false);
  const [authError, setAuthError] = useState("");
  const [linksError, setLinksError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function bootstrap() {
      setIsInitializing(true);
      setAuthError("");

      try {
        const nextSession = await getSession();
        if (cancelled) {
          return;
        }

        setSession(nextSession);
        await loadLinks(setLinks, setLinksError, setIsLoadingLinks);
        navigate("/links", { replace: true });
      } catch (error) {
        if (cancelled) {
          return;
        }

        if (!isUnauthorizedError(error)) {
          setAuthError(getErrorMessage(error, "会话校验失败"));
        }

        setSession(null);
        setLinks([]);
        navigate("/login", { replace: true });
      } finally {
        if (!cancelled) {
          setIsInitializing(false);
        }
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, [navigate]);

  async function handleLogin(username: string, password: string) {
    setIsAuthenticating(true);
    setAuthError("");

    try {
      const nextSession = await login(username, password);
      setSession(nextSession);
      await loadLinks(setLinks, setLinksError, setIsLoadingLinks);
      navigate("/links", { replace: true });
    } catch (error) {
      setAuthError(getErrorMessage(error, "登录失败"));
    } finally {
      setIsAuthenticating(false);
    }
  }

  async function handleReload() {
    await loadLinks(setLinks, setLinksError, setIsLoadingLinks);
  }

  async function handleLogout() {
    try {
      await logout();
    } catch {
      // Logout should always clear local state even if the request fails.
    }

    setSession(null);
    setLinks([]);
    setLinksError("");
    setAuthError("");
    navigate("/login", { replace: true });
  }

  if (isInitializing) {
    return (
      <main className="login-shell">
        <section className="login-card">
          <p className="eyebrow">shorturl Admin</p>
          <h1>加载后台会话</h1>
          <p className="muted-copy">正在校验当前登录状态并同步短链列表。</p>
        </section>
      </main>
    );
  }

  return (
    <Routes>
      <Route
        path="/login"
        element={<LoginPage isSubmitting={isAuthenticating} error={authError} onSubmit={handleLogin} />}
      />
      <Route
        path="/links"
        element={
          <ProtectedRoute session={session}>
            <LinksPage
              session={session as AuthSession}
              links={links}
              isLoading={isLoadingLinks}
              error={linksError}
              onReload={handleReload}
              onLogout={handleLogout}
            />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to={session ? "/links" : "/login"} replace />} />
    </Routes>
  );
}

async function loadLinks(
  setLinks: (items: Link[]) => void,
  setLinksError: (message: string) => void,
  setIsLoadingLinks: (loading: boolean) => void,
) {
  setIsLoadingLinks(true);
  setLinksError("");

  try {
    const nextLinks = await listLinks();
    setLinks(nextLinks);
  } catch (error) {
    setLinksError(getErrorMessage(error, "列表加载失败"));
  } finally {
    setIsLoadingLinks(false);
  }
}
