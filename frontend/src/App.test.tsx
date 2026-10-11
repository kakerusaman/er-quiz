import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import type { Question, Summary } from "./types";

const summaries: Summary[] = [{ id: 7, title: "ECサイトの注文", blankCount: 1 }];

const question: Question = {
  id: 7,
  title: "ECサイトの注文",
  scenario: "会員が商品を注文する。",
  entities: [
    {
      id: 1,
      name: "会員",
      x: 0,
      y: 0,
      blank: null,
      attributes: [
        { id: 1, name: "会員番号", isPk: true, isFk: false, blank: null },
        { id: 2, name: null, isPk: false, isFk: false, blank: { id: 10, symbol: "イ" } },
      ],
    },
  ],
  relationships: [],
  choices: [
    { id: 100, label: "メールアドレス" },
    { id: 101, label: "在庫数" },
  ],
};

function mockFetch() {
  return vi.fn(async (url: string, init?: RequestInit) => {
    if (init?.method === "POST") {
      expect(url).toBe("/api/questions/7/answers");
      return Response.json({ correct: { 10: true }, score: 1, total: 1 });
    }
    if (url === "/api/questions") return Response.json(summaries);
    expect(url).toBe("/api/questions/7");
    return Response.json(question);
  });
}

describe("App", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("空欄を選んで答え合わせすると点数が出る", async () => {
    const fetch = mockFetch();
    vi.stubGlobal("fetch", fetch);
    render(<App />);

    const select = await screen.findByLabelText("空欄 イ");
    await userEvent.selectOptions(select, "メールアドレス");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));

    expect(await screen.findByText("1 / 1 問正解")).toBeInTheDocument();
    const post = fetch.mock.calls.find(([, init]) => init?.method === "POST")!;
    expect(JSON.parse(post[1]!.body as string)).toEqual({ answers: { 10: 100 } });
  });

  it("問題が0件なら知らせる", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => Response.json([])),
    );
    render(<App />);
    expect(await screen.findByRole("alert")).toHaveTextContent("問題がまだありません");
  });
});
