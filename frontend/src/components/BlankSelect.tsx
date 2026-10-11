import type { BlankRef, Choice } from "../types";

type Props = {
  blank: BlankRef;
  choices: Choice[];
  value?: number;
  correct?: boolean;
  onChange: (blankId: number, choiceId: number) => void;
};

// 空欄の枠。破線の枠に記号（ア、イ…）を出し、選んだ選択肢を置くとその文字列に変わる。
// ステップ5でドラッグ＆ドロップに置き換えるまでは、枠の中のプルダウンで選ぶ。
export function BlankSelect({ blank, choices, value, correct, onChange }: Props) {
  const state = correct === undefined ? "" : correct ? "correct" : "wrong";
  return (
    <select
      className={`nodrag blank ${value === undefined ? "empty" : "filled"} ${state}`}
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
