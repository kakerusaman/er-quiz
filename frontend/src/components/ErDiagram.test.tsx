import { render } from "@testing-library/react";
import { Position } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import type { Cardinality, Entity } from "../types";
import { CardinalityMark } from "./CardinalityMark";
import { connectingSides } from "./ErDiagram";

const at = (x: number, y: number): Entity => ({ id: 1, name: "E", x, y, blank: null, attributes: [] });

describe("connectingSides", () => {
  it("横に離れていれば左右、縦に離れていれば上下の辺をつなぐ", () => {
    expect(connectingSides(at(0, 0), at(340, 0))).toEqual(["right", "left"]);
    expect(connectingSides(at(340, 0), at(0, 0))).toEqual(["left", "right"]);
    expect(connectingSides(at(680, 0), at(680, 240))).toEqual(["bottom", "top"]);
    expect(connectingSides(at(680, 240), at(680, 0))).toEqual(["top", "bottom"]);
  });
});

describe("CardinalityMark", () => {
  // 鳥の足は線2本、｜は線1本、○は円1つ。
  it.each<[Cardinality, number, number]>([
    ["1", 2, 0],
    ["0..1", 1, 1],
    ["1..N", 3, 0],
    ["0..N", 2, 1],
  ])("%s は線 %i 本と円 %i 個で描く", (cardinality, lines, circles) => {
    const { container } = render(
      <svg>
        <CardinalityMark x={0} y={0} position={Position.Right} cardinality={cardinality} />
      </svg>,
    );
    expect(container.querySelectorAll("line")).toHaveLength(lines);
    expect(container.querySelectorAll("circle")).toHaveLength(circles);
  });

  it("線が出ていく向きに合わせて回す", () => {
    const { container } = render(
      <svg>
        <CardinalityMark x={10} y={20} position={Position.Top} cardinality="1" />
      </svg>,
    );
    expect(container.querySelector("g")).toHaveAttribute("transform", "translate(10 20) rotate(-90)");
  });
});
