import type { Blank } from "../types";

type Props = {
  blank: Blank;
  value: string;
  correct?: boolean;
  onChange: (blankId: string, value: string) => void;
};

// 空欄1つ分の選択肢。採点後は正誤で色が変わる。
export function BlankSelect({ blank, value, correct, onChange }: Props) {
  const state = correct === undefined ? "" : correct ? "correct" : "wrong";
  return (
    <select
      className={`nodrag blank ${state}`}
      aria-label={`空欄 ${blank.id}`}
      value={value}
      onChange={(e) => onChange(blank.id, e.target.value)}
    >
      <option value="">？？？</option>
      {blank.choices.map((c) => (
        <option key={c} value={c}>
          {c}
        </option>
      ))}
    </select>
  );
}
