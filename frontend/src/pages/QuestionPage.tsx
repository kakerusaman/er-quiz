import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { useEffect, useState } from "react";
import { applyDrop, availableChoices, type DragSource } from "../answers";
import { fetchQuestion, submitAnswers } from "../api";
import { collisionDetection } from "../collision";
import { ChoicePanel } from "../components/ChoicePanel";
import { ErDiagram } from "../components/ErDiagram";
import { Header } from "../components/Header";
import { LoadError } from "../components/LoadError";
import { ScenarioPanel } from "../components/ScenarioPanel";
import { navigate } from "../router";
import type { Answers, AttemptResult, Question } from "../types";

// 問題画面（/questions/{id}）。上から業務の説明文、ER図、解答群（答え合わせ後は点数と解説）。
export function QuestionPage({ id }: { id: number }) {
  const [question, setQuestion] = useState<Question>();
  const [failed, setFailed] = useState(false);
  const [answers, setAnswers] = useState<Answers>({});
  const [result, setResult] = useState<AttemptResult>();
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState(false);
  const [dragging, setDragging] = useState<string>();

  // 少し動かしてからドラッグを始める（クリックだけでは始まらない）。
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor),
  );

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

  const label = (choiceId: number) => question?.choices.find((c) => c.id === choiceId)?.label;

  const onDragStart = ({ active }: DragStartEvent) => {
    const source = active.data.current as DragSource | undefined;
    setDragging(source && label(source.choiceId));
  };

  const onDragEnd = ({ active, over }: DragEndEvent) => {
    setDragging(undefined);
    const source = active.data.current as DragSource | undefined;
    if (!source) return;
    const toBlankId = over?.data.current?.blankId as number | undefined;
    setAnswers((prev) => applyDrop(prev, source, toBlankId));
  };

  const onSubmit = async () => {
    if (!question) return;
    setSubmitting(true);
    setSubmitError(false);
    try {
      setResult(await submitAnswers(question.id, answers));
    } catch {
      setSubmitError(true);
    } finally {
      setSubmitting(false);
    }
  };

  const retry = () => {
    setAnswers({});
    setResult(undefined);
  };

  // 解答中に空欄を埋めていたら、一覧へ戻る前に確認する。
  const leave = () => {
    const filled = !result && Object.keys(answers).length > 0;
    if (filled && !window.confirm("解答が消えますがよいですか？")) return;
    navigate("/");
  };

  if (failed) return <LoadError actionLabel="一覧へ戻る" onAction={() => navigate("/")} />;

  return (
    <>
      <Header>
        <button className="link" onClick={leave}>
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
          {result && (
            <section className="card score-card">
              <p className="score">
                {result.total}問中 {result.score}問正解
              </p>
            </section>
          )}
          <ScenarioPanel scenario={question.scenario} />
          <DndContext
            sensors={sensors}
            collisionDetection={collisionDetection}
            onDragStart={onDragStart}
            onDragEnd={onDragEnd}
            onDragCancel={() => setDragging(undefined)}
          >
            <ErDiagram question={question} answers={answers} result={result} />
            {result ? (
              <section className="card explanation" aria-labelledby="explanation-title">
                <h2 id="explanation-title">解説</h2>
                <p>{result.explanation}</p>
                <div className="choices-actions">
                  <button className="secondary" onClick={retry}>
                    もう一度解く
                  </button>
                  <button className="primary" onClick={() => navigate("/")}>
                    一覧へ戻る
                  </button>
                </div>
              </section>
            ) : (
              <ChoicePanel choices={availableChoices(question.choices, answers)}>
                {submitError && (
                  <span className="error-text" role="alert">
                    送信に失敗しました。もう一度お試しください
                  </span>
                )}
                <button className="primary" onClick={onSubmit} disabled={submitting}>
                  答え合わせ
                </button>
              </ChoicePanel>
            )}
            <DragOverlay dropAnimation={null}>
              {dragging && <span className="choice-chip overlay">{dragging}</span>}
            </DragOverlay>
          </DndContext>
        </main>
      )}
    </>
  );
}
