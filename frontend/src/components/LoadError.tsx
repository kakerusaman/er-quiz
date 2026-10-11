import { Header } from "./Header";

type Props = {
  actionLabel: string;
  onAction: () => void;
};

// 読み込みエラーの画面。一覧なら [再読み込み]、問題画面なら [一覧へ戻る] を出す。
export function LoadError({ actionLabel, onAction }: Props) {
  return (
    <>
      <Header />
      <main className="container">
        <div className="card message" role="alert">
          <p>問題を読み込めませんでした</p>
          <button className="primary" onClick={onAction}>
            {actionLabel}
          </button>
        </div>
      </main>
    </>
  );
}
