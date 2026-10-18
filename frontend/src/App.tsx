import { useEffect, useMemo, useState } from "react";
import { fetchMe, logout } from "./api";
import { AuthContext, type AuthState } from "./auth";
import { LoadError } from "./components/LoadError";
import { LoginPage } from "./pages/LoginPage";
import { QuestionList } from "./pages/QuestionList";
import { QuestionPage } from "./pages/QuestionPage";
import { SignupPage } from "./pages/SignupPage";
import { navigate, questionIdFromPath, usePath } from "./router";
import type { User } from "./types";
import { useIsNarrow } from "./useIsNarrow";

export function App() {
  const path = usePath();
  const narrow = useIsNarrow();
  const id = questionIdFromPath(path);

  // ログイン状態。開いたときに /api/me で確かめる。確かめられなければ未ログインとして扱う。
  const [user, setUser] = useState<User | null>();
  useEffect(() => {
    let active = true;
    fetchMe().then(
      (u) => active && setUser(u),
      () => active && setUser(null),
    );
    return () => {
      active = false;
    };
  }, []);
  const auth = useMemo<AuthState>(
    () => ({
      user,
      setUser,
      logout: async () => {
        await logout();
        setUser(null);
      },
    }),
    [user],
  );

  const page =
    path === "/" ? (
      <QuestionList />
    ) : path === "/login" ? (
      <LoginPage />
    ) : path === "/signup" ? (
      <SignupPage />
    ) : id !== undefined ? (
      <QuestionPage key={id} id={id} />
    ) : (
      <LoadError actionLabel="一覧へ戻る" onAction={() => navigate("/")} />
    );

  // 狭い画面では案内だけを見せる。画面は隠すだけで残すので、幅を戻すと解答中の状態に戻る。
  return (
    <AuthContext.Provider value={auth}>
      {narrow && <p className="narrow-notice">PCで開いてください</p>}
      <div className="app" hidden={narrow}>
        {page}
      </div>
    </AuthContext.Provider>
  );
}
