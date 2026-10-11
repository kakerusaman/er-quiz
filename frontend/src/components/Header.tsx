import type { ReactNode } from "react";
import { navigate } from "../router";

// 画面上部の白いヘッダー。children を渡さなければアプリ名を出す。
export function Header({ children }: { children?: ReactNode }) {
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
      </div>
    </header>
  );
}
