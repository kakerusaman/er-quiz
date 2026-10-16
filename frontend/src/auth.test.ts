import { describe, expect, it } from "vitest";
import { ApiError } from "./api";
import { authErrorMessage, nextPath } from "./auth";

describe("authErrorMessage", () => {
  it("API のエラーを日本語の文にする", () => {
    expect(authErrorMessage(new ApiError(401, "invalid credentials"))).toBe("ログインIDかパスワードが違います");
    expect(authErrorMessage(new ApiError(409, "login id taken"))).toBe("このログインIDはすでに使われています");
  });

  it("知らないエラーや通信の失敗は、もう一度試すよう案内する", () => {
    const retry = "送信できませんでした。時間をおいてもう一度試してください";
    expect(authErrorMessage(new ApiError(500, "internal error"))).toBe(retry);
    expect(authErrorMessage(new TypeError("Failed to fetch"))).toBe(retry);
  });
});

describe("nextPath", () => {
  it("next のパスへ戻る", () => {
    expect(nextPath("?next=%2Fquestions%2F7")).toBe("/questions/7");
  });

  it("next がない・ほかのサイトを指すときは一覧へ戻る", () => {
    expect(nextPath("")).toBe("/");
    expect(nextPath("?next=https%3A%2F%2Fevil.example")).toBe("/");
    expect(nextPath("?next=%2F%2Fevil.example")).toBe("/");
  });
});
