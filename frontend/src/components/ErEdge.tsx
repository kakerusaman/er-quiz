import { BaseEdge, getSmoothStepPath, type Edge, type EdgeProps } from "@xyflow/react";
import type { Cardinality } from "../types";
import { CardinalityMark } from "./CardinalityMark";

export type ErEdgeType = Edge<{ sourceCardinality: Cardinality; targetCardinality: Cardinality }, "er">;

// リレーションシップの線。エンティティの辺から垂直に出し、両端に多重度の記号を描く。
export function ErEdge(props: EdgeProps<ErEdgeType>) {
  const { sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, data } = props;
  const [path] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 0,
    offset: 30,
  });
  return (
    <>
      <BaseEdge path={path} className="er-edge" />
      {data && (
        <>
          <CardinalityMark x={sourceX} y={sourceY} position={sourcePosition} cardinality={data.sourceCardinality} />
          <CardinalityMark x={targetX} y={targetY} position={targetPosition} cardinality={data.targetCardinality} />
        </>
      )}
    </>
  );
}
