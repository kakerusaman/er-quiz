import { describe, expect, it } from "vitest";
import { applyDrop } from "./answers";

describe("applyDrop", () => {
  it("解答群から空欄へ置く", () => {
    expect(applyDrop({}, { choiceId: 5 }, 1)).toEqual({ 1: 5 });
  });

  it("同じ選択肢を何度でも使える", () => {
    expect(applyDrop({ 1: 5 }, { choiceId: 5 }, 2)).toEqual({ 1: 5, 2: 5 });
  });

  it("選択肢が入った空欄へ別の選択肢を置くと入れ替わる", () => {
    expect(applyDrop({ 1: 5 }, { choiceId: 6 }, 1)).toEqual({ 1: 6 });
  });

  it("空欄から空の空欄へ動かす", () => {
    expect(applyDrop({ 1: 5 }, { choiceId: 5, fromBlankId: 1 }, 2)).toEqual({ 2: 5 });
  });

  it("空欄から埋まった空欄へ動かすと中身が入れ替わる", () => {
    expect(applyDrop({ 1: 5, 2: 6 }, { choiceId: 5, fromBlankId: 1 }, 2)).toEqual({ 1: 6, 2: 5 });
  });

  it("空欄の外へ出すと空に戻る", () => {
    expect(applyDrop({ 1: 5, 2: 6 }, { choiceId: 5, fromBlankId: 1 }, undefined)).toEqual({ 2: 6 });
  });

  it("解答群の選択肢を空欄の外へ落としても何も変わらない", () => {
    const answers = { 1: 5 };
    expect(applyDrop(answers, { choiceId: 6 }, undefined)).toEqual({ 1: 5 });
  });

  it("同じ空欄に落とし直しても変わらない", () => {
    const answers = { 1: 5 };
    expect(applyDrop(answers, { choiceId: 5, fromBlankId: 1 }, 1)).toBe(answers);
  });
});
