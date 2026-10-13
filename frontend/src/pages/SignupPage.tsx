import { useState, type FormEvent } from "react";
import { signup } from "../api";
import { authErrorMessage, nextPath, useAuth } from "../auth";
import { Header } from "../components/Header";
import { navigate } from "../router";

// 新規登録画面（/signup）。登録するとそのままログインし、?next= の画面（なければ一覧）へ戻る。
export function SignupPage() {
  const { setUser } = useAuth();
  const [loginId, setLoginId] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string>();
  const [sending, setSending] = useState(false);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setSending(true);
    setError(undefined);
    try {
      setUser(await signup(loginId, email, password));
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
        <section className="card auth-card" aria-labelledby="signup-title">
          <h1 id="signup-title">新規登録</h1>
          <form onSubmit={submit}>
            <label>
              ログインID
              <input value={loginId} onChange={(e) => setLoginId(e.target.value)} autoComplete="username" required />
              <span className="hint">半角英数字と _ の3〜20文字。画面に表示されます</span>
            </label>
            <label>
              メールアドレス
              <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" required />
            </label>
            <label>
              パスワード
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
                required
              />
              <span className="hint">8〜72文字</span>
            </label>
            {error && (
              <p className="error-text" role="alert">
                {error}
              </p>
            )}
            <button className="primary" type="submit" disabled={sending}>
              登録する
            </button>
          </form>
          <p className="auth-switch">
            アカウントがある場合は{" "}
            <button className="link" onClick={() => navigate(`/login${window.location.search}`)}>
              ログイン
            </button>
          </p>
        </section>
      </main>
    </>
  );
}
