import { useCallback, useEffect, useState } from "react";
import { fetchQuestion, fetchQuestions, submitAnswers } from "./api";
import { ErDiagram } from "./components/ErDiagram";
import type { Answers, GradeResult, Question } from "./types";

export function App() {
  const [question, setQuestion] = useState<Question>();
  const [answers, setAnswers] = useState<Answers>({});
  const [result, setResult] = useState<GradeResult>();
  const [error, setError] = useState<string>();

  // 問題一覧画面ができるまでは、一覧の先頭の問題を表示する。
  useEffect(() => {
    fetchQuestions()
      .then((list) => {
        if (list.length === 0) throw new Error("問題がまだありません");
        return fetchQuestion(list[0].id);
      })
      .then(setQuestion, (e: Error) => setError(e.message));
  }, []);

  const onAnswer = useCallback((blankId: number, choiceId: number) => {
    setAnswers((prev) => ({ ...prev, [blankId]: choiceId }));
    setResult(undefined);
  }, []);

  const onSubmit = async () => {
    try {
      if (!question) return;
      setResult(await submitAnswers(question.id, answers));
    } catch (e) {
      setError((e as Error).message);
    }
  };

  if (error) return <p role="alert">読み込みに失敗しました: {error}</p>;
  if (!question) return <p>読み込み中…</p>;

  return (
    <main>
      <header>
        <h1>{question.title}</h1>
        <button onClick={onSubmit}>答え合わせ</button>
        {result && (
          <span className="score">
            {result.score} / {result.total} 問正解
          </span>
        )}
      </header>
      <ErDiagram question={question} answers={answers} result={result} onAnswer={onAnswer} />
    </main>
  );
}
