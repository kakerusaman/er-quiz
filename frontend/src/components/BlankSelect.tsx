import type { BlankRef, Choice } from "../types";

type Props = {
  blank: BlankRef;
  choices: Choice[];
  value?: number;
  correct?: boolean;
  onChange: (blankId: number, choiceId: number) => void;
};

// 空欄1つ分の選択肢。採点後は正誤で色が変わる。
// ドラッグ＆ドロップに置き換えるまでの仮の部品。
export function BlankSelect({ blank, choices, value, correct, onChange }: Props) {
  const state = correct === undefined ? "" : correct ? "correct" : "wrong";
  return (
    <select
      className={`nodrag blank ${state}`}
      aria-label={`空欄 ${blank.symbol}`}
      value={value ?? ""}
      onChange={(e) => onChange(blank.id, Number(e.target.value))}
    >
      <option value="">{blank.symbol}</option>
      {choices.map((c) => (
        <option key={c.id} value={c.id}>
          {c.label}
        </option>
      ))}
    </select>
  );
}
