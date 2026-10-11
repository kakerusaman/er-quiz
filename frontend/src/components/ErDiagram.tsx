import { useEffect, useMemo } from "react";
import { Panel, ReactFlow, useNodesState, useReactFlow } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import type { Answers, AttemptResult, Entity, Question } from "../types";
import { ErEdge, type ErEdgeType } from "./ErEdge";
import { TableNode, type Side, type TableNodeType } from "./TableNode";

const nodeTypes = { table: TableNode };
const edgeTypes = { er: ErEdge };

type Props = {
  question: Question;
  answers: Answers;
  result?: AttemptResult;
};

/**
 * 2つのエンティティの位置から、線をつなぐ辺を決める。
 * 横に離れていれば左右の辺、縦に離れていれば上下の辺をつなぐ。
 */
export function connectingSides(from: Entity, to: Entity): [Side, Side] {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  if (Math.abs(dx) >= Math.abs(dy)) return dx >= 0 ? ["right", "left"] : ["left", "right"];
  return dy >= 0 ? ["bottom", "top"] : ["top", "bottom"];
}

// 図の右下に置く拡大・縮小ボタン。
function ZoomButtons() {
  const { zoomIn, zoomOut } = useReactFlow();
  return (
    <Panel position="bottom-right" className="zoom-buttons">
      <button aria-label="拡大" onClick={() => zoomIn()}>
        ＋
      </button>
      <button aria-label="縮小" onClick={() => zoomOut()}>
        －
      </button>
    </Panel>
  );
}

// 問題を React Flow のノードとエッジに変換して表示する。
// エンティティの配置は固定で、図全体の拡大・縮小と移動だけができる。
export function ErDiagram({ question, answers, result }: Props) {
  const [nodes, setNodes] = useNodesState<TableNodeType>([]);

  // 問題が変わったらノードを作り直す。
  useEffect(() => {
    setNodes(
      question.entities.map((entity) => ({
        id: String(entity.id),
        type: "table",
        position: { x: entity.x, y: entity.y },
        data: { entity, choices: question.choices, answers: {} },
      })),
    );
  }, [question, setNodes]);

  // 回答や採点結果が変わったら、計測済みのサイズは保ったまま data だけ差し替える。
  useEffect(() => {
    setNodes((ns) => ns.map((n) => ({ ...n, data: { ...n.data, answers, result } })));
  }, [answers, result, setNodes]);

  const edges: ErEdgeType[] = useMemo(() => {
    const byId = new Map(question.entities.map((e) => [e.id, e]));
    return question.relationships.flatMap((r) => {
      const from = byId.get(r.fromEntityId);
      const to = byId.get(r.toEntityId);
      if (!from || !to) return [];
      const [sourceHandle, targetHandle] = connectingSides(from, to);
      return [
        {
          id: String(r.id),
          type: "er" as const,
          source: String(r.fromEntityId),
          target: String(r.toEntityId),
          sourceHandle,
          targetHandle,
          data: { sourceCardinality: r.fromCardinality, targetCardinality: r.toCardinality },
        },
      ];
    });
  }, [question]);

  return (
    <div className="card diagram">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        fitView
        fitViewOptions={{ padding: 0.15, maxZoom: 1 }}
        minZoom={0.3}
        maxZoom={2}
        attributionPosition="bottom-left"
      >
        <ZoomButtons />
      </ReactFlow>
    </div>
  );
}
