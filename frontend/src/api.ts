import type { Answers, AttemptResult, Question, Summary } from "./types";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init);
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json() as Promise<T>;
}

export function fetchQuestions(): Promise<Summary[]> {
  return request("/api/questions");
}

export function fetchQuestion(id: number): Promise<Question> {
  return request(`/api/questions/${id}`);
}

export function submitAnswers(id: number, answers: Answers): Promise<AttemptResult> {
  return request(`/api/questions/${id}/attempts`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ answers }),
  });
}
