import { createContext, useContext } from "react";
import { ApiError } from "./api";
import type { User } from "./types";

export type AuthState = {
  /** 確かめ中は undefined、未ログインは null */
  user: User | null | undefined;
  setUser: (user: User | null) => void;
  /** ログアウトする。失敗したら reject する。 */
  logout: () => Promise<void>;
};

export const AuthContext = createContext<AuthState>({
  user: undefined,
  setUser: () => {},
  logout: async () => {},
});

/** ログイン状態。App の AuthContext.Provider の中で使う。 */
export function useAuth(): AuthState {
  return useContext(AuthContext);
}

const messages: Record<string, string> = {
  "invalid credentials": "ログインIDかパスワードが違います",
  "invalid login id": "ログインIDは半角英数字と _ の3〜20文字で入力してください",
  "invalid email": "メールアドレスを正しく入力してください",
  "invalid password": "パスワードは8〜72文字で入力してください",
  "login id taken": "このログインIDはすでに使われています",
  "email taken": "このメールアドレスはすでに登録されています",
};

/** ログイン・新規登録の失敗を、画面に出す文に直す。 */
export function authErrorMessage(e: unknown): string {
  const code = e instanceof ApiError ? e.code : undefined;
  return (code && messages[code]) || "送信できませんでした。時間をおいてもう一度試してください";
}

/** ログイン後に戻る先。`?next=` が同じサイトのパスでなければ一覧へ戻る。 */
export function nextPath(search: string): string {
  const next = new URLSearchParams(search).get("next");
  return next && next.startsWith("/") && !next.startsWith("//") ? next : "/";
}
