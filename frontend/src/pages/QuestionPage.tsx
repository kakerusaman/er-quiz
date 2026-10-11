import { useCallback, useEffect, useState } from "react";
import { fetchQuestion, submitAnswers } from "../api";
import { ErDiagram } from "../components/ErDiagram";
import { Header } from "../components/Header";
import { LoadError } from "../components/LoadError";
import { navigate } from "../router";
import type { Answers, AttemptResult, Question } from "../types";

// 問題画面（/questions/{id}）。見た目と答え合わせ後の表示はステップ4・5で作る。
export function QuestionPage({ id }: { id: number }) {
  const [question, setQuestion] = useState<Question>();
  const [failed, setFailed] = useState(false);
  const [answers, setAnswers] = useState<Answers>({});
  const [result, setResult] = useState<AttemptResult>();
  const [submitError, setSubmitError] = useState(false);

  useEffect(() => {
    let active = true;
    fetchQuestion(id).then(
      (q) => active && setQuestion(q),
      () => active && setFailed(true),
    );
    return () => {
      active = false;
    };
  }, [id]);

  const onAnswer = useCallback((blankId: number, choiceId: number) => {
    setAnswers((prev) => ({ ...prev, [blankId]: choiceId }));
    setResult(undefined);
  }, []);

  const onSubmit = async () => {
    if (!question) return;
    try {
      setSubmitError(false);
      setResult(await submitAnswers(question.id, answers));
    } catch {
      setSubmitError(true);
    }
  };

  if (failed) return <LoadError actionLabel="一覧へ戻る" onAction={() => navigate("/")} />;

  return (
    <>
      <Header>
        <button className="link" onClick={() => navigate("/")}>
          ← 一覧へ
        </button>
        {question && <h1 className="page-title">{question.title}</h1>}
      </Header>
      {!question ? (
        <main className="container">
          <p>読み込み中…</p>
        </main>
      ) : (
        <main className="container question-page">
          <div className="toolbar">
            <button className="primary" onClick={onSubmit}>
              答え合わせ
            </button>
            {result && (
              <span className="score">
                {result.score} / {result.total} 問正解
              </span>
            )}
            {submitError && <span role="alert">送信に失敗しました。もう一度お試しください</span>}
          </div>
          <ErDiagram question={question} answers={answers} result={result} onAnswer={onAnswer} />
        </main>
      )}
    </>
  );
}
