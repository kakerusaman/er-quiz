import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import type { Answers, BlankRef, Choice, Entity, AttemptResult } from "../types";
import { BlankSelect } from "./BlankSelect";

export type TableNodeData = {
  entity: Entity;
  choices: Choice[];
  answers: Answers;
  result?: AttemptResult;
  onAnswer: (blankId: number, choiceId: number) => void;
};

export type TableNodeType = Node<TableNodeData, "table">;

// ER図上のエンティティ1つ。名前が空欄の箇所は選択肢で表示する。
export function TableNode({ data }: NodeProps<TableNodeType>) {
  const { entity, choices, answers, result, onAnswer } = data;
  const renderName = (name: string | null, blank: BlankRef | null) =>
    blank ? (
      <BlankSelect
        blank={blank}
        choices={choices}
        value={answers[blank.id]}
        correct={result?.results.find((r) => r.blankId === blank.id)?.correct}
        onChange={onAnswer}
      />
    ) : (
      <span>{name}</span>
    );

  return (
    <div className="table-node">
      <Handle type="target" position={Position.Left} />
      <div className="table-name">{renderName(entity.name, entity.blank)}</div>
      <ul>
        {entity.attributes.map((a) => (
          <li key={a.id}>
            <span className="key">{[a.isPk && "PK", a.isFk && "FK"].filter(Boolean).join("/")}</span>
            {renderName(a.name, a.blank)}
          </li>
        ))}
      </ul>
      <Handle type="source" position={Position.Right} />
    </div>
  );
}
