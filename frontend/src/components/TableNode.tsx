import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import type { Answers, Blank, GradeResult, Table } from "../types";
import { BlankSelect } from "./BlankSelect";

export type TableNodeData = {
  table: Table;
  blanks: Record<string, Blank>;
  answers: Answers;
  result?: GradeResult;
  onAnswer: (blankId: string, value: string) => void;
};

export type TableNodeType = Node<TableNodeData, "table">;

// ER図上のテーブル1つ。名前が空欄の箇所は選択肢で表示する。
export function TableNode({ data }: NodeProps<TableNodeType>) {
  const { table, blanks, answers, result, onAnswer } = data;
  const renderName = (name: string | undefined, blankId: string | undefined) =>
    blankId && blanks[blankId] ? (
      <BlankSelect
        blank={blanks[blankId]}
        value={answers[blankId] ?? ""}
        correct={result?.correct[blankId]}
        onChange={onAnswer}
      />
    ) : (
      <span>{name}</span>
    );

  return (
    <div className="table-node">
      <Handle type="target" position={Position.Left} />
      <div className="table-name">{renderName(table.name, table.blankId)}</div>
      <ul>
        {table.columns.map((c) => (
          <li key={c.id}>
            <span className="key">{c.pk ? "PK" : c.fk ? "FK" : ""}</span>
            {renderName(c.name, c.blankId)}
            <span className="type">{c.type}</span>
          </li>
        ))}
      </ul>
      <Handle type="source" position={Position.Right} />
    </div>
  );
}
