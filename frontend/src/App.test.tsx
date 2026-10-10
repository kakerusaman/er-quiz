import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import type { Question } from "./types";

const question: Question = {
  id: "shop",
  title: "ECサイトの注文",
  tables: [
    {
      id: "users",
      name: "users",
      x: 0,
      y: 0,
      columns: [{ id: "users.email", type: "varchar", pk: false, fk: false, blankId: "b1" }],
    },
  ],
  relations: [],
  blanks: [{ id: "b1", choices: ["email", "price"] }],
};

function mockFetch() {
  return vi.fn(async (url: string, init?: RequestInit) => {
    if (init?.method === "POST") {
      return Response.json({ correct: { b1: true }, score: 1, total: 1 });
    }
    expect(url).toBe("/api/questions/shop");
    return Response.json(question);
  });
}

describe("App", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("空欄を選んで答え合わせすると点数が出る", async () => {
    const fetch = mockFetch();
    vi.stubGlobal("fetch", fetch);
    render(<App />);

    const select = await screen.findByLabelText("空欄 b1");
    await userEvent.selectOptions(select, "email");
    await userEvent.click(screen.getByRole("button", { name: "答え合わせ" }));

    expect(await screen.findByText("1 / 1 問正解")).toBeInTheDocument();
    const body = JSON.parse(fetch.mock.calls[1][1]!.body as string);
    expect(body).toEqual({ answers: { b1: "email" } });
  });
});
