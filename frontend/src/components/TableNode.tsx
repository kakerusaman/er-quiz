import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import type { Answers, AttemptResult, BlankRef, Choice, Entity } from "../types";
import { BlankSelect } from "./BlankSelect";

export type TableNodeData = {
  entity: Entity;
  choices: Choice[];
  answers: Answers;
  result?: AttemptResult;
  onAnswer: (blankId: number, choiceId: number) => void;
};

export type TableNodeType = Node<TableNodeData, "table">;

/** 線をつなぐ辺。ハンドルの id にも使う。 */
export const sides = [
  ["top", Position.Top],
  ["right", Position.Right],
  ["bottom", Position.Bottom],
  ["left", Position.Left],
] as const;

export type Side = (typeof sides)[number][0];

// ER図上のエンティティ1つ。見出しと属性の行を並べ、空欄の箇所は空欄の枠を出す。
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
      {sides.map(([id, position]) => (
        <span key={id}>
          <Handle type="source" id={id} position={position} isConnectable={false} />
          <Handle type="target" id={id} position={position} isConnectable={false} />
        </span>
      ))}
      <div className="table-name">{renderName(entity.name, entity.blank)}</div>
      <ul>
        {entity.attributes.map((a) => {
          const key = [a.isPk && "PK", a.isFk && "FK"].filter(Boolean).join("/");
          return (
            <li key={a.id}>
              {key && <span className="key-badge">{key}</span>}
              {renderName(a.name, a.blank)}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
