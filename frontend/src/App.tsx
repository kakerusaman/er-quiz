import { LoadError } from "./components/LoadError";
import { QuestionList } from "./pages/QuestionList";
import { QuestionPage } from "./pages/QuestionPage";
import { navigate, questionIdFromPath, usePath } from "./router";
import { useIsNarrow } from "./useIsNarrow";

export function App() {
  const path = usePath();
  const narrow = useIsNarrow();
  const id = questionIdFromPath(path);

  const page =
    path === "/" ? (
      <QuestionList />
    ) : id !== undefined ? (
      <QuestionPage key={id} id={id} />
    ) : (
      <LoadError actionLabel="一覧へ戻る" onAction={() => navigate("/")} />
    );

  // 狭い画面では案内だけを見せる。画面は隠すだけで残すので、幅を戻すと解答中の状態に戻る。
  return (
    <>
      {narrow && <p className="narrow-notice">PCで開いてください</p>}
      <div className="app" hidden={narrow}>
        {page}
      </div>
    </>
  );
}
