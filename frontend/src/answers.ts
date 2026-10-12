import type { Answers, Choice } from "./types";

/** ドラッグしているもの。解答群の選択肢なら fromBlankId はない。 */
export type DragSource = { choiceId: number; fromBlankId?: number };

/**
 * ドロップしたあとの回答を返す。toBlankId は落とした空欄（空欄の外なら undefined）。
 * - 解答群から空欄へ: その空欄に入る（入っていた選択肢は解答群に戻る）。
 * - 空欄から別の空欄へ: 2つの空欄の中身が入れ替わる。
 * - 空欄から空欄の外へ: 元の空欄が空に戻る。
 */
export function applyDrop(answers: Answers, source: DragSource, toBlankId: number | undefined): Answers {
  const { choiceId, fromBlankId } = source;
  if (toBlankId === fromBlankId) return answers;
  const next = { ...answers };
  if (fromBlankId === undefined && toBlankId !== undefined) {
    // 選択肢は1つの空欄にしか置けない。ほかの空欄に残っていたら外す。
    for (const [blankId, id] of Object.entries(next)) if (id === choiceId) delete next[Number(blankId)];
  }
  if (fromBlankId !== undefined) {
    const displaced = toBlankId === undefined ? undefined : answers[toBlankId];
    if (displaced === undefined) delete next[fromBlankId];
    else next[fromBlankId] = displaced;
  }
  if (toBlankId !== undefined) next[toBlankId] = choiceId;
  return next;
}

/** 解答群に出す選択肢。空欄に置いたものは除き、API の順を保つ（外すと元の位置に戻る）。 */
export function availableChoices(choices: Choice[], answers: Answers): Choice[] {
  const used = new Set(Object.values(answers));
  return choices.filter((c) => !used.has(c.id));
}
