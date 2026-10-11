import { useDraggable } from "@dnd-kit/core";
import type { ReactNode } from "react";
import type { DragSource } from "../answers";
import type { Choice } from "../types";

type Props = {
  choices: Choice[];
  /** 右下に置くボタンなど */
  children?: ReactNode;
};

// 画面下部の解答群。選択肢を API の順に並べ、空欄へドラッグできる。
export function ChoicePanel({ choices, children }: Props) {
  return (
    <section className="card choices" aria-labelledby="choices-title">
      <div className="choices-head">
        <h2 id="choices-title">解答群</h2>
        <span className="hint">選択肢を空欄にドラッグして置きます。同じ選択肢は何度でも使えます。</span>
      </div>
      <ul className="choice-list">
        {choices.map((c) => (
          <li key={c.id}>
            <ChoiceChip choice={c} />
          </li>
        ))}
      </ul>
      <div className="choices-actions">{children}</div>
    </section>
  );
}

function ChoiceChip({ choice }: { choice: Choice }) {
  const data: DragSource = { choiceId: choice.id };
  const { setNodeRef, attributes, listeners, isDragging } = useDraggable({ id: `choice-${choice.id}`, data });
  return (
    <span
      ref={setNodeRef}
      {...attributes}
      {...listeners}
      className={`choice-chip${isDragging ? " dragging" : ""}`}
      aria-label={`選択肢 ${choice.label}`}
    >
      {choice.label}
    </span>
  );
}
