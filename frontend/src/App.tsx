import { useCallback, useEffect, useState } from "react";
import { fetchQuestion, submitAnswers } from "./api";
import { ErDiagram } from "./components/ErDiagram";
import type { Answers, GradeResult, Question } from "./types";

const QUESTION_ID = "shop";

export function App() {
  const [question, setQuestion] = useState<Question>();
  const [answers, setAnswers] = useState<Answers>({});
  const [result, setResult] = useState<GradeResult>();
  const [error, setError] = useState<string>();

  useEffect(() => {
    fetchQuestion(QUESTION_ID).then(setQuestion, (e: Error) => setError(e.message));
  }, []);

  const onAnswer = useCallback((blankId: string, value: string) => {
    setAnswers((prev) => ({ ...prev, [blankId]: value }));
    setResult(undefined);
  }, []);

  const onSubmit = async () => {
    try {
      setResult(await submitAnswers(QUESTION_ID, answers));
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
