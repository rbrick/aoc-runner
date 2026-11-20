import { useEffect, useState } from "react";
import { ApiError, getSession, login, logout, type Session } from "../api";

type Props = {
  onLogin: () => void;
  onLogout: () => void;
};

export function AuthStatus({ onLogin, onLogout }: Props) {
  const [session, setSession] = useState<Session | null>(null);
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getSession()
      .then(setSession)
      .catch((err: ApiError) => setError(err.message));
  }, []);

  async function tryLogin() {
    setChecking(true);
    setError(null);

    try {
      const result = await login();
      setSession(result);

      if (result.logged_in) {
        onLogin();
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setChecking(false);
    }
  }

  async function tryLogout() {
    setChecking(true);
    setError(null);

    try {
      setSession(await logout());
      onLogout();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setChecking(false);
    }
  }

  if (session?.logged_in) {
    return (
      <div className="auth">
        {error && <p className="auth-hint auth-error">{error}</p>}
        <span className="pill pill-ok">Logged in</span>
        <button className="button button-small" onClick={tryLogout} disabled={checking}>
          {checking ? "Logging out…" : "Log out"}
        </button>
      </div>
    );
  }

  return (
    <div className="auth">
      {session?.message && (
        <p className="auth-hint">
          No session found. <a href={session.login_url} target="_blank" rel="noreferrer">Log in to Advent of Code</a>,
          then try again.
        </p>
      )}
      {error && <p className="auth-hint auth-error">{error}</p>}
      <button className="button" onClick={tryLogin} disabled={checking}>
        {checking ? "Checking…" : session?.message ? "Try again" : "Log in"}
      </button>
    </div>
  );
}
