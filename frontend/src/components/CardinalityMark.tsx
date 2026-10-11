import { Position } from "@xyflow/react";
import type { Cardinality } from "../types";

// 線がエンティティから出ていく向き（右・下・左・上）を角度にしたもの。
const angles: Record<Position, number> = {
  [Position.Right]: 0,
  [Position.Bottom]: 90,
  [Position.Left]: 180,
  [Position.Top]: -90,
};

type Props = {
  /** 線とエンティティの接点 */
  x: number;
  y: number;
  /** 接点がエンティティのどの辺にあるか */
  position: Position;
  cardinality: Cardinality;
};

// IE記法の線の端に付ける多重度の記号。
// 「1」は｜｜、「0..1」は○｜、「1..N」は｜と鳥の足、「0..N」は○と鳥の足。
// 座標は接点を原点にし、x 軸の正の向きをエンティティから離れる向きにして描く。
export function CardinalityMark({ x, y, position, cardinality }: Props) {
  const many = cardinality.endsWith("N");
  const zero = cardinality.startsWith("0");
  const bar = (at: number) => <line x1={at} y1={-7} x2={at} y2={7} />;
  return (
    <g
      className="cardinality-mark"
      data-cardinality={cardinality}
      transform={`translate(${x} ${y}) rotate(${angles[position]})`}
    >
      {many ? (
        <>
          <line x1={12} y1={0} x2={0} y2={-8} />
          <line x1={12} y1={0} x2={0} y2={8} />
        </>
      ) : (
        bar(8)
      )}
      {zero ? <circle cx={17} cy={0} r={4.5} /> : bar(many ? 16 : 13)}
    </g>
  );
}
