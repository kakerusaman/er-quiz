import { useEffect, useMemo } from "react";
import { Background, Controls, ReactFlow, useNodesState, type Edge } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import type { Answers, AttemptResult, Question } from "../types";
import { TableNode, type TableNodeType } from "./TableNode";

const nodeTypes = { table: TableNode };

type Props = {
  question: Question;
  answers: Answers;
  result?: AttemptResult;
  onAnswer: (blankId: number, choiceId: number) => void;
};

// 問題をReact Flowのノードとエッジに変換して表示する。テーブルはドラッグで動かせる。
export function ErDiagram({ question, answers, result, onAnswer }: Props) {
  const [nodes, setNodes, onNodesChange] = useNodesState<TableNodeType>([]);

  // 問題が変わったら初期位置でノードを作り直す。
  useEffect(() => {
    setNodes(
      question.entities.map((entity) => ({
        id: String(entity.id),
        type: "table",
        position: { x: entity.x, y: entity.y },
        data: { entity, choices: question.choices, answers: {}, onAnswer },
      })),
    );
  }, [question, setNodes]);

  // 回答や採点結果が変わったら、位置やサイズは保ったまま data だけ差し替える。
  useEffect(() => {
    setNodes((ns) => ns.map((n) => ({ ...n, data: { ...n.data, answers, result, onAnswer } })));
  }, [answers, result, onAnswer, setNodes]);

  const edges: Edge[] = useMemo(
    () =>
      question.relationships.map((r) => ({
        id: String(r.id),
        source: String(r.fromEntityId),
        target: String(r.toEntityId),
        // IE記法の線はステップ4で描く。それまでは多重度を文字で出す。
        label: `${r.fromCardinality} : ${r.toCardinality}`,
      })),
    [question],
  );

  return (
    <div className="diagram">
      <ReactFlow nodes={nodes} edges={edges} nodeTypes={nodeTypes} onNodesChange={onNodesChange} fitView>
        <Background />
        <Controls />
      </ReactFlow>
    </div>
  );
}
