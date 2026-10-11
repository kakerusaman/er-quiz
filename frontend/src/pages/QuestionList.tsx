import { useEffect, useState } from "react";
import { fetchQuestions } from "../api";
import { Header } from "../components/Header";
import { LoadError } from "../components/LoadError";
import { navigate } from "../router";
import type { Summary } from "../types";

// 問題一覧（/）。新しい問題が上（API の返す順のまま）。
export function QuestionList() {
  const [list, setList] = useState<Summary[]>();
  const [failed, setFailed] = useState(false);
  const [reload, setReload] = useState(0);

  useEffect(() => {
    let active = true;
    setFailed(false);
    setList(undefined);
    fetchQuestions().then(
      (l) => active && setList(l),
      () => active && setFailed(true),
    );
    return () => {
      active = false;
    };
  }, [reload]);

  if (failed) return <LoadError actionLabel="再読み込み" onAction={() => setReload((n) => n + 1)} />;

  return (
    <>
      <Header />
      <main className="container">
        <h2 className="section-title">問題一覧</h2>
        {!list ? (
          <p>読み込み中…</p>
        ) : list.length === 0 ? (
          <div className="card message">
            <p>問題がまだありません</p>
          </div>
        ) : (
          <ul className="card question-list">
            {list.map((q) => (
              <li key={q.id}>
                <span className="question-title">{q.title}</span>
                <span className="blank-count">空欄 {q.blankCount}</span>
                <button
                  className="primary"
                  aria-label={`${q.title}を解く`}
                  onClick={() => navigate(`/questions/${q.id}`)}
                >
                  解く
                </button>
              </li>
            ))}
          </ul>
        )}
      </main>
    </>
  );
}
