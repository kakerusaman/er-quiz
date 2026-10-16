import type { ReactNode } from "react";
import { useAuth } from "../auth";
import { confirmLeave, navigate } from "../router";

type Props = {
  children?: ReactNode;
  /** false ならログイン状態を出さない（ログイン・新規登録の画面） */
  account?: boolean;
};

// 画面上部の白いヘッダー。children を渡さなければアプリ名を出す。右端にはログイン状態を出す。
export function Header({ children, account = true }: Props) {
  return (
    <header className="app-header">
      <div className="container header-inner">
        {children ?? (
          <a
            className="app-name"
            href="/"
            onClick={(e) => {
              e.preventDefault();
              navigate("/");
            }}
          >
            ER図クイズ
          </a>
        )}
        {account && <UserMenu />}
      </div>
    </header>
  );
}

// ログイン中ならログインIDと [ログアウト]、未ログインなら [ログイン] [新規登録] を出す。確かめ中は何も出さない。
function UserMenu() {
  const { user, logout } = useAuth();
  if (user === undefined) return null;

  if (user === null) {
    // ログイン・登録の後に、今の画面へ戻れるようにする。解答中なら移る前に確認する。
    const go = (path: string) => {
      if (confirmLeave()) navigate(`${path}?next=${encodeURIComponent(window.location.pathname)}`);
    };
    return (
      <nav className="user-menu" aria-label="アカウント">
        <button className="link" onClick={() => go("/login")}>
          ログイン
        </button>
        <button className="secondary small" onClick={() => go("/signup")}>
          新規登録
        </button>
      </nav>
    );
  }

  return (
    <nav className="user-menu" aria-label="アカウント">
      <span className="login-id">{user.loginId}</span>
      <button className="link" onClick={() => logout().catch(() => window.alert("ログアウトできませんでした"))}>
        ログアウト
      </button>
    </nav>
  );
}
