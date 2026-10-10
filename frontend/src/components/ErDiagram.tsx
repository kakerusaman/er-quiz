import { useEffect, useMemo } from "react";
import { Background, Controls, ReactFlow, useNodesState, type Edge } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import type { Answers, GradeResult, Question } from "../types";
import { TableNode, type TableNodeType } from "./TableNode";

const nodeTypes = { table: TableNode };

type Props = {
  question: Question;
  answers: Answers;
  result?: GradeResult;
  onAnswer: (blankId: string, value: string) => void;
};

// 問題をReact Flowのノードとエッジに変換して表示する。テーブルはドラッグで動かせる。
export function ErDiagram({ question, answers, result, onAnswer }: Props) {
  const blanks = useMemo(
    () => Object.fromEntries(question.blanks.map((b) => [b.id, b])),
    [question],
  );
  const [nodes, setNodes, onNodesChange] = useNodesState<TableNodeType>([]);

  // 問題が変わったら初期位置でノードを作り直す。
  useEffect(() => {
    setNodes(
      question.tables.map((table) => ({
        id: table.id,
        type: "table",
        position: { x: table.x, y: table.y },
        data: { table, blanks, answers: {}, onAnswer },
      })),
    );
  }, [question, blanks, setNodes]);

  // 回答や採点結果が変わったら、位置やサイズは保ったまま data だけ差し替える。
  useEffect(() => {
    setNodes((ns) => ns.map((n) => ({ ...n, data: { ...n.data, answers, result, onAnswer } })));
  }, [answers, result, onAnswer, setNodes]);

  const edges: Edge[] = useMemo(
    () =>
      question.relations.map((r) => ({
        id: r.id,
        source: r.from,
        target: r.to,
        label: r.cardinality,
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
