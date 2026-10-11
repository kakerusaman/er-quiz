import { useDraggable, useDroppable } from "@dnd-kit/core";
import type { DragSource } from "../answers";
import type { BlankRef, BlankResult } from "../types";

type Props = {
  blank: BlankRef;
  /** 置いた選択肢 */
  placed?: { id: number; label: string };
  /** 答え合わせ後の結果と正解の文字列 */
  result?: { blank: BlankResult; answerLabel: string };
};

// ER図の空欄。解答中は選択肢を受け取る枠で、置いた選択肢は別の空欄や外へドラッグできる。
// 答え合わせ後は正誤のチップになり、不正解なら下に正解を出す。
export function BlankSlot({ blank, placed, result }: Props) {
  if (result) {
    const { correct } = result.blank;
    return (
      <span className="blank-result">
        <span
          className={`blank-chip ${correct ? "correct" : "wrong"}`}
          aria-label={`空欄 ${blank.symbol}: ${placed?.label ?? "未回答"}、${correct ? "正解" : "不正解"}`}
        >
          {placed?.label}
          <span className="result-icon" aria-hidden="true">
            {correct ? "✓" : "✕"}
          </span>
        </span>
        {!correct && <span className="correct-answer">正解: {result.answerLabel}</span>}
      </span>
    );
  }
  return <DroppableBlank blank={blank} placed={placed} />;
}

function DroppableBlank({ blank, placed }: Omit<Props, "result">) {
  const drop = useDroppable({ id: `blank-${blank.id}`, data: { blankId: blank.id } });
  const source: DragSource | undefined = placed && { choiceId: placed.id, fromBlankId: blank.id };
  const drag = useDraggable({ id: `placed-${blank.id}`, data: source, disabled: !placed });
  return (
    <span
      ref={(el) => {
        drop.setNodeRef(el);
        drag.setNodeRef(el);
      }}
      {...drag.attributes}
      {...drag.listeners}
      // nodrag / nopan: 空欄のドラッグで ER図が動かないようにする（React Flow のクラス）
      className={[
        "nodrag nopan blank",
        placed ? "filled" : "empty",
        drop.isOver && "over",
        drag.isDragging && "dragging",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-label={`空欄 ${blank.symbol}`}
      data-placed={placed?.label}
    >
      {placed ? placed.label : blank.symbol}
    </span>
  );
}
