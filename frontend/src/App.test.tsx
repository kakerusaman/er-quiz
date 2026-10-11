import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import type { Question, Summary } from "./types";

const summaries: Summary[] = [
  { id: 8, title: "図書館の貸出", blankCount: 5 },
  { id: 7, title: "ECサイトの注文", blankCount: 1 },
];

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
      expect(url).toBe("/api/questions/7/attempts");
      return Response.json({
        attemptId: 1,
        score: 1,
        total: 1,
        results: [{ blankId: 10, correct: true, answerChoiceId: 100 }],
        explanation: "会員はメールアドレスを持つ。",
      });
    }
    if (url === "/api/questions") return Response.json(summaries);
    if (url === "/api/questions/7") return Response.json(question);
    return Response.json({ error: "question not found" }, { status: 404 });
  });
}

/** matchMedia の代わり。setNarrow で画面幅の変化を起こせる。 */
function mockMatchMedia(narrow: boolean) {
  const listeners = new Set<() => void>();
  const mql = {
    get matches() {
      return narrow;
    },
    addEventListener: (_: string, l: () => void) => listeners.add(l),
    removeEventListener: (_: string, l: () => void) => listeners.delete(l),
  };
  vi.stubGlobal("matchMedia", () => mql);
  return (next: boolean) => {
    narrow = next;
    act(() => listeners.forEach((l) => l()));
  };
}

describe("App", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/");
    mockMatchMedia(false);
  });
  afterEach(() => vi.unstubAllGlobals());

  it("一覧を新しい順に出し、[解く] で問題画面へ移る", async () => {
    vi.stubGlobal("fetch", mockFetch());
    render(<App />);

    const items = await screen.findAllByRole("listitem");
    expect(items.map((li) => li.textContent)).toEqual(["図書館の貸出空欄 5解く", "ECサイトの注文空欄 1解く"]);

    await userEvent.click(screen.getByRole("button", { name: "ECサイトの注文を解く" }));
    expect(window.location.pathname).toBe("/questions/7");
    expect(await screen.findByRole("heading", { name: "ECサイトの注文" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "← 一覧へ" }));
    expect(window.location.pathname).toBe("/");
    expect(await screen.findByText("問題一覧")).toBeInTheDocument();
  });

  it("問題画面で空欄を選んで答え合わせすると点数が出る", async () => {
    window.history.replaceState(null, "", "/questions/7");
    const fetch = mockFetch();
    vi.stubGlobal("fetch", fetch);
    render(<App />);

    const select = await screen.findByLabelText("空欄 イ");
    await userEvent.selectOptions(select, "メールアドレス");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));

    expect(await screen.findByText("1 / 1 問正解")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByLabelText("空欄 イ")).toHaveClass("correct"));
    const post = fetch.mock.calls.find(([, init]) => init?.method === "POST")!;
    expect(JSON.parse(post[1]!.body as string)).toEqual({ answers: { 10: 100 } });
  });

  it("問題が0件なら知らせる", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => Response.json([])),
    );
    render(<App />);
    expect(await screen.findByText("問題がまだありません")).toBeInTheDocument();
  });

  it("問題画面に説明文・解答群・拡大縮小ボタンを出し、説明文はたためる", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    render(<App />);

    const scenario = await screen.findByText("会員が商品を注文する。");
    expect(scenario).toBeVisible();
    const title = screen.getByRole("button", { name: /業務の説明文/ });
    expect(title).toHaveAttribute("aria-expanded", "true");

    await userEvent.click(screen.getByRole("button", { name: "小さくする" }));
    expect(scenario).not.toBeVisible();
    expect(title).toHaveAttribute("aria-expanded", "false");
    await userEvent.click(screen.getByRole("button", { name: "広げる" }));
    expect(scenario).toBeVisible();

    const choices = within(screen.getByRole("region", { name: "解答群" })).getAllByRole("listitem");
    expect(choices.map((c) => c.textContent)).toEqual(["メールアドレス", "在庫数"]);
    expect(screen.getByRole("button", { name: "拡大" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "縮小" })).toBeInTheDocument();
  });

  it("一覧の取得に失敗したら [再読み込み] でやり直せる", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 500 }))
      .mockResolvedValue(Response.json(summaries));
    vi.stubGlobal("fetch", fetch);
    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent("問題を読み込めませんでした");
    await userEvent.click(screen.getByRole("button", { name: "再読み込み" }));
    expect(await screen.findByText("図書館の貸出")).toBeInTheDocument();
  });

  it("存在しない問題は読み込みエラーになり、[一覧へ戻る] で一覧へ戻る", async () => {
    window.history.replaceState(null, "", "/questions/999");
    vi.stubGlobal("fetch", mockFetch());
    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent("問題を読み込めませんでした");
    await userEvent.click(screen.getByRole("button", { name: "一覧へ戻る" }));
    expect(window.location.pathname).toBe("/");
    expect(await screen.findByText("図書館の貸出")).toBeInTheDocument();
  });

  it("1024px 未満では案内だけを出し、幅を戻すと解答が残っている", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    const setNarrow = mockMatchMedia(false);
    render(<App />);

    await userEvent.selectOptions(await screen.findByLabelText("空欄 イ"), "メールアドレス");

    setNarrow(true);
    expect(screen.getByText("PCで開いてください")).toBeVisible();
    expect(screen.queryByRole("button", { name: "答え合わせ" })).not.toBeInTheDocument();

    setNarrow(false);
    expect(screen.queryByText("PCで開いてください")).not.toBeInTheDocument();
    expect(screen.getByLabelText("空欄 イ")).toHaveValue("100");
  });
});
