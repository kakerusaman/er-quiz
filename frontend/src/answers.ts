import type { Answers } from "./types";

/** ドラッグしているもの。解答群の選択肢なら fromBlankId はない。 */
export type DragSource = { choiceId: number; fromBlankId?: number };

/**
 * ドロップしたあとの回答を返す。toBlankId は落とした空欄（空欄の外なら undefined）。
 * - 解答群から空欄へ: その空欄に入る（入っていた選択肢と入れ替わる）。
 * - 空欄から別の空欄へ: 2つの空欄の中身が入れ替わる。
 * - 空欄から空欄の外へ: 元の空欄が空に戻る。
 */
export function applyDrop(answers: Answers, source: DragSource, toBlankId: number | undefined): Answers {
  const { choiceId, fromBlankId } = source;
  if (toBlankId === fromBlankId) return answers;
  const next = { ...answers };
  if (fromBlankId !== undefined) {
    const displaced = toBlankId === undefined ? undefined : answers[toBlankId];
    if (displaced === undefined) delete next[fromBlankId];
    else next[fromBlankId] = displaced;
  }
  if (toBlankId !== undefined) next[toBlankId] = choiceId;
  return next;
}
