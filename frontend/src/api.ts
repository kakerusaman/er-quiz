import type { Answers, AttemptResult, Question, Summary, User } from "./types";

/** API がエラーを返したとき。code はレスポンスの `{"error": "..."}`（読めなければ undefined）。 */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string | undefined,
  ) {
    super(`${status} ${code ?? ""}`.trim());
  }
}

async function send(path: string, init?: RequestInit): Promise<Response> {
  const res = await fetch(path, init);
  if (!res.ok) {
    const body: unknown = await res.json().catch(() => undefined);
    const code =
      typeof body === "object" && body !== null && "error" in body && typeof body.error === "string"
        ? body.error
        : undefined;
    throw new ApiError(res.status, code);
  }
  return res;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await send(path, init);
  return res.json() as Promise<T>;
}

function postJSON(body: unknown): RequestInit {
  return { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) };
}

export function fetchQuestions(): Promise<Summary[]> {
  return request("/api/questions");
}

export function fetchQuestion(id: number): Promise<Question> {
  return request(`/api/questions/${id}`);
}

export function submitAnswers(id: number, answers: Answers): Promise<AttemptResult> {
  return request(`/api/questions/${id}/attempts`, postJSON({ answers }));
}

/** ログイン中のユーザー。未ログインなら null。 */
export async function fetchMe(): Promise<User | null> {
  try {
    return await request<User>("/api/me");
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) return null;
    throw e;
  }
}

export function signup(loginId: string, email: string, password: string): Promise<User> {
  return request("/api/signup", postJSON({ loginId, email, password }));
}

export function login(loginId: string, password: string): Promise<User> {
  return request("/api/login", postJSON({ loginId, password }));
}

export async function logout(): Promise<void> {
  await send("/api/logout", postJSON({}));
}
