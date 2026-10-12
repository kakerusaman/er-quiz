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
      const { answers } = JSON.parse(init.body as string) as { answers: Record<string, number> };
      const correct = answers[10] === 100;
      return Response.json({
        attemptId: 1,
        score: correct ? 1 : 0,
        total: 1,
        results: [{ blankId: 10, correct, answerChoiceId: 100 }],
        explanation: "会員はメールアドレスを持つ。",
      });
    }
    if (url === "/api/questions") return Response.json(summaries);
    if (url === "/api/questions/7") return Response.json(question);
    return Response.json({ error: "question not found" }, { status: 404 });
  });
}

/** キーボードで選択肢をつかみ、一番近い空欄（この問題では空欄イ）に置く。 */
async function dropChoice(label: string) {
  screen.getByRole("button", { name: `選択肢 ${label}` }).focus();
  await userEvent.keyboard("[Space]");
  await userEvent.keyboard("[Space]");
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

  it("選択肢を空欄に置いて答え合わせすると、点数・正誤・解説が出る", async () => {
    window.history.replaceState(null, "", "/questions/7");
    const fetch = mockFetch();
    vi.stubGlobal("fetch", fetch);
    render(<App />);

    expect(await screen.findByLabelText("空欄 イ")).toHaveTextContent("イ");
    await dropChoice("メールアドレス");
    await waitFor(() => expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("メールアドレス"));

    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));
    expect(await screen.findByText("1問中 1問正解")).toBeInTheDocument();
    expect(await screen.findByLabelText("空欄 イ: メールアドレス、正解")).toBeInTheDocument();
    expect(screen.queryByText(/^正解:/)).not.toBeInTheDocument();
    expect(screen.getByRole("region", { name: "解説" })).toHaveTextContent("会員はメールアドレスを持つ。");
    expect(screen.queryByRole("region", { name: "解答群" })).not.toBeInTheDocument();

    const post = fetch.mock.calls.find(([, init]) => init?.method === "POST")!;
    expect(JSON.parse(post[1]!.body as string)).toEqual({ answers: { 10: 100 } });
  });

  it("選択肢をつかんで空欄に重ねると、その空欄が強調される", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    render(<App />);

    const blank = await screen.findByLabelText("空欄 イ");
    expect(blank).not.toHaveClass("over");
    screen.getByRole("button", { name: "選択肢 メールアドレス" }).focus();
    await userEvent.keyboard("[Space]");
    await waitFor(() => expect(blank).toHaveClass("over"));
    await userEvent.keyboard("[Escape]");
    await waitFor(() => expect(blank).not.toHaveClass("over"));
  });

  it("不正解と未回答には ✕ と正解を出し、[もう一度解く] で空欄が空に戻る", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    render(<App />);

    await screen.findByLabelText("空欄 イ");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));
    expect(await screen.findByText("1問中 0問正解")).toBeInTheDocument();
    expect(await screen.findByLabelText("空欄 イ: 未回答、不正解")).toBeInTheDocument();
    expect(screen.getByText("正解: メールアドレス")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "もう一度解く" }));
    expect(await screen.findByLabelText("空欄 イ")).toHaveTextContent("イ");
    expect(screen.queryByText("1問中 0問正解")).not.toBeInTheDocument();
    expect(screen.getByRole("region", { name: "解答群" })).toBeInTheDocument();

    // もう一度解いた結果は不正解の表示を残さない
    await dropChoice("在庫数");
    await waitFor(() => expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("在庫数"));
  });

  it("空欄を埋めてから [← 一覧へ] を押すと確認を出す", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    const confirm = vi.fn(() => false);
    vi.stubGlobal("confirm", confirm);
    render(<App />);

    await screen.findByLabelText("空欄 イ");
    await dropChoice("メールアドレス");
    await waitFor(() => expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("メールアドレス"));

    await userEvent.click(screen.getByRole("button", { name: "← 一覧へ" }));
    expect(confirm).toHaveBeenCalledWith("解答が消えますがよいですか？");
    expect(window.location.pathname).toBe("/questions/7");

    confirm.mockReturnValue(true);
    await userEvent.click(screen.getByRole("button", { name: "← 一覧へ" }));
    expect(window.location.pathname).toBe("/");
  });

  it("答え合わせ後の [一覧へ戻る] は確認を出さない", async () => {
    window.history.replaceState(null, "", "/questions/7");
    vi.stubGlobal("fetch", mockFetch());
    const confirm = vi.fn(() => false);
    vi.stubGlobal("confirm", confirm);
    render(<App />);

    await screen.findByLabelText("空欄 イ");
    await dropChoice("メールアドレス");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));
    await userEvent.click(await screen.findByRole("button", { name: "一覧へ戻る" }));
    expect(confirm).not.toHaveBeenCalled();
    expect(window.location.pathname).toBe("/");
  });

  it("答え合わせの送信に失敗したら、解答を残して知らせる", async () => {
    window.history.replaceState(null, "", "/questions/7");
    const ok = mockFetch();
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) =>
        init?.method === "POST" ? new Response(null, { status: 500 }) : ok(url, init),
      ),
    );
    render(<App />);

    await screen.findByLabelText("空欄 イ");
    await dropChoice("メールアドレス");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("送信に失敗しました。もう一度お試しください");
    expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("メールアドレス");
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

    await screen.findByLabelText("空欄 イ");
    await dropChoice("メールアドレス");
    await waitFor(() => expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("メールアドレス"));

    setNarrow(true);
    expect(screen.getByText("PCで開いてください")).toBeVisible();
    expect(screen.queryByRole("button", { name: "答え合わせ" })).not.toBeInTheDocument();

    setNarrow(false);
    expect(screen.queryByText("PCで開いてください")).not.toBeInTheDocument();
    expect(screen.getByLabelText("空欄 イ")).toHaveTextContent("メールアドレス");
  });
});
