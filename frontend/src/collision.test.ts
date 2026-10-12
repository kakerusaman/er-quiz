import type { ClientRect, DroppableContainer } from "@dnd-kit/core";
import { describe, expect, it } from "vitest";
import { collisionDetection } from "./collision";

function rect(left: number, top: number, width: number, height: number): ClientRect {
  return { left, top, width, height, right: left + width, bottom: top + height };
}

// 空欄ア (0,0)-(100,30)、空欄イ (200,0)-(300,30)
const blanks = new Map([
  ["blank-1", rect(0, 0, 100, 30)],
  ["blank-2", rect(200, 0, 100, 30)],
]);

function detect(chip: ClientRect, pointer: { x: number; y: number } | null) {
  const collisions = collisionDetection({
    active: {} as never,
    collisionRect: chip,
    droppableRects: blanks,
    droppableContainers: [...blanks.keys()].map((id) => ({ id }) as unknown as DroppableContainer),
    pointerCoordinates: pointer,
  });
  return collisions.map((c) => c.id);
}

describe("collisionDetection", () => {
  it("カーソルが空欄の外でも、チップが空欄に重なれば当たる", () => {
    // チップ (80,20)-(180,60) の左上が空欄アの右下に重なる。カーソルはチップの中央で、空欄の外
    expect(detect(rect(80, 20, 100, 40), { x: 130, y: 40 })).toEqual(["blank-1"]);
  });

  it("カーソルが空欄の中でも、チップが重なっていなければ当たらない", () => {
    expect(detect(rect(400, 100, 100, 40), { x: 50, y: 15 })).toEqual([]);
  });

  it("2つの空欄に重なったら、重なりが大きいほうが先", () => {
    // チップ (70,0)-(210,30): 空欄アと 30px、空欄イと 10px 重なる
    expect(detect(rect(70, 0, 140, 30), { x: 140, y: 15 })).toEqual(["blank-1", "blank-2"]);
    // チップ (90,0)-(230,30): 空欄アと 10px、空欄イと 30px 重なる
    expect(detect(rect(90, 0, 140, 30), { x: 160, y: 15 })).toEqual(["blank-2", "blank-1"]);
  });

  it("キーボード操作では一番近い空欄", () => {
    expect(detect(rect(220, 100, 60, 30), null)[0]).toBe("blank-2");
  });
});
