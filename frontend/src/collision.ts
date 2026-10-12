import { closestCenter, rectIntersection, type CollisionDetection } from "@dnd-kit/core";

/**
 * どの空欄に重なっているかの判定。
 * マウスでは、つかんでいる選択肢のチップと空欄の枠が重なった空欄（複数なら重なりが大きいほう）。
 * キーボード操作では一番近い空欄。
 */
export const collisionDetection: CollisionDetection = (args) =>
  args.pointerCoordinates ? rectIntersection(args) : closestCenter(args);
