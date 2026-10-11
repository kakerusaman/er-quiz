import type { ReactNode } from "react";
import type { Choice } from "../types";

type Props = {
  choices: Choice[];
  /** 右下に置くボタンなど */
  children?: ReactNode;
};

// 画面下部の解答群。選択肢を API の順に並べる（ドラッグはステップ5で付ける）。
export function ChoicePanel({ choices, children }: Props) {
  return (
    <section className="card choices" aria-labelledby="choices-title">
      <div className="choices-head">
        <h2 id="choices-title">解答群</h2>
        <span className="hint">選択肢を空欄にドラッグして置きます。同じ選択肢は何度でも使えます。</span>
      </div>
      <ul className="choice-list">
        {choices.map((c) => (
          <li key={c.id} className="choice-chip">
            {c.label}
          </li>
        ))}
      </ul>
      <div className="choices-actions">{children}</div>
    </section>
  );
}
