import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "../App";
import type { Summary, User } from "../types";

const summaries: Summary[] = [{ id: 7, title: "ECサイトの注文", blankCount: 1 }];
const taro: User = { id: 1, loginId: "taro", email: "taro@example.com" };

/** サーバーの代わり。loggedIn でログイン状態を持ち、ログイン・ログアウトで切り替える。 */
function mockServer(loggedIn: boolean) {
  const fetch = vi.fn(async (url: string, init?: RequestInit) => {
    if (url === "/api/me") return loggedIn ? Response.json(taro) : Response.json({ error: "login required" }, { status: 401 });
    if (url === "/api/login") {
      const { loginId, password } = JSON.parse(init?.body as string) as { loginId: string; password: string };
      if (loginId !== "taro" || password !== "password") {
        return Response.json({ error: "invalid credentials" }, { status: 401 });
      }
      loggedIn = true;
      return Response.json(taro);
    }
    if (url === "/api/logout") {
      loggedIn = false;
      return new Response(null, { status: 204 });
    }
    if (url === "/api/questions") return Response.json(summaries);
    return Response.json({ error: "question not found" }, { status: 404 });
  });
  vi.stubGlobal("fetch", fetch);
  return fetch;
}

async function accountMenu() {
  return within(await screen.findByRole("navigation", { name: "アカウント" }));
}

describe("LoginPage", () => {
  beforeEach(() => window.history.replaceState(null, "", "/"));
  afterEach(() => vi.unstubAllGlobals());

  it("未ログインならヘッダーに [ログイン] を出し、ログインすると元の画面に戻ってログインIDを出す", async () => {
    mockServer(false);
    render(<App />);
    await screen.findByText("ECサイトの注文");

    await userEvent.click((await accountMenu()).getByRole("button", { name: "ログイン" }));
    expect(window.location.pathname).toBe("/login");
    await userEvent.type(screen.getByLabelText("ログインID"), "taro");
    await userEvent.type(screen.getByLabelText("パスワード"), "password");
    await userEvent.click(screen.getByRole("button", { name: "ログイン" }));

    expect(await screen.findByText("ECサイトの注文")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
    const menu = await accountMenu();
    expect(menu.getByText("taro")).toBeInTheDocument();
    expect(menu.queryByRole("button", { name: "ログイン" })).not.toBeInTheDocument();
  });

  it("ログインIDかパスワードが違えば知らせ、画面に留まる", async () => {
    mockServer(false);
    window.history.replaceState(null, "", "/login");
    render(<App />);

    await userEvent.type(screen.getByLabelText("ログインID"), "taro");
    await userEvent.type(screen.getByLabelText("パスワード"), "wrong-password");
    await userEvent.click(screen.getByRole("button", { name: "ログイン" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("ログインIDかパスワードが違います");
    expect(window.location.pathname).toBe("/login");
  });

  it("[ログアウト] で未ログインの表示に戻る", async () => {
    const fetch = mockServer(true);
    render(<App />);

    await userEvent.click((await accountMenu()).getByRole("button", { name: "ログアウト" }));

    const menu = await accountMenu();
    expect(await menu.findByRole("button", { name: "ログイン" })).toBeInTheDocument();
    expect(menu.queryByText("taro")).not.toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith("/api/logout", expect.objectContaining({ method: "POST" }));
  });
});
