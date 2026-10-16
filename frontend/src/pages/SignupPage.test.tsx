import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "../App";
import type { Question } from "../types";

const question: Question = {
  id: 7,
  title: "ECサイトの注文",
  scenario: "会員が商品を注文する。",
  entities: [],
  relationships: [],
  choices: [],
};

/** サーバーの代わり。登録済みのログインIDは taken。 */
function mockServer() {
  const fetch = vi.fn(async (url: string, init?: RequestInit) => {
    if (url === "/api/me") return Response.json({ error: "login required" }, { status: 401 });
    if (url === "/api/signup") {
      const body = JSON.parse(init?.body as string) as { loginId: string; email: string; password: string };
      if (body.loginId === "taken") return Response.json({ error: "login id taken" }, { status: 409 });
      return Response.json({ id: 2, loginId: body.loginId, email: body.email }, { status: 201 });
    }
    if (url === "/api/questions/7") return Response.json(question);
    return Response.json({ error: "question not found" }, { status: 404 });
  });
  vi.stubGlobal("fetch", fetch);
  return fetch;
}

async function fillForm(loginId: string) {
  await userEvent.type(screen.getByLabelText(/ログインID/), loginId);
  await userEvent.type(screen.getByLabelText("メールアドレス"), "hanako@example.com");
  await userEvent.type(screen.getByLabelText(/パスワード/), "password");
  await userEvent.click(screen.getByRole("button", { name: "登録する" }));
}

describe("SignupPage", () => {
  beforeEach(() => window.history.replaceState(null, "", "/"));
  afterEach(() => vi.unstubAllGlobals());

  it("問題画面から新規登録すると、ログインした状態で問題画面に戻る", async () => {
    const fetch = mockServer();
    window.history.replaceState(null, "", "/questions/7");
    render(<App />);
    const menu = within(await screen.findByRole("navigation", { name: "アカウント" }));

    await userEvent.click(await menu.findByRole("button", { name: "新規登録" }));
    expect(window.location.pathname).toBe("/signup");
    await fillForm("hanako");

    expect(await screen.findByRole("heading", { name: "ECサイトの注文" })).toBeInTheDocument();
    expect(window.location.pathname).toBe("/questions/7");
    expect(within(screen.getByRole("navigation", { name: "アカウント" })).getByText("hanako")).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith(
      "/api/signup",
      expect.objectContaining({
        body: JSON.stringify({ loginId: "hanako", email: "hanako@example.com", password: "password" }),
      }),
    );
  });

  it("使われているログインIDなら知らせる", async () => {
    mockServer();
    window.history.replaceState(null, "", "/signup");
    render(<App />);

    await fillForm("taken");

    expect(await screen.findByRole("alert")).toHaveTextContent("このログインIDはすでに使われています");
    expect(window.location.pathname).toBe("/signup");
  });
});
