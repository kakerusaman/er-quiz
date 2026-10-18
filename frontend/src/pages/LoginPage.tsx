import { useState, type FormEvent } from "react";
import { login } from "../api";
import { authErrorMessage, nextPath, useAuth } from "../auth";
import { Header } from "../components/Header";
import { navigate } from "../router";

// ログイン画面（/login）。ログインできたら ?next= の画面（なければ一覧）へ戻る。
export function LoginPage() {
  const { setUser } = useAuth();
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string>();
  const [sending, setSending] = useState(false);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setSending(true);
    setError(undefined);
    try {
      setUser(await login(loginId, password));
      navigate(nextPath(window.location.search));
    } catch (err) {
      setError(authErrorMessage(err));
      setSending(false);
    }
  };

  return (
    <>
      <Header account={false} />
      <main className="container">
        <section className="card auth-card" aria-labelledby="login-title">
          <h1 id="login-title">ログイン</h1>
          <form onSubmit={submit}>
            <label>
              ログインID
              <input value={loginId} onChange={(e) => setLoginId(e.target.value)} autoComplete="username" required />
            </label>
            <label>
              パスワード
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
                required
              />
            </label>
            {error && (
              <p className="error-text" role="alert">
                {error}
              </p>
            )}
            <button className="primary" type="submit" disabled={sending}>
              ログイン
            </button>
          </form>
          <p className="auth-switch">
            アカウントがない場合は{" "}
            <button className="link" onClick={() => navigate(`/signup${window.location.search}`)}>
              新規登録
            </button>
          </p>
        </section>
      </main>
    </>
  );
}
