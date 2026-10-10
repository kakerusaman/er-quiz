import type { Answers, GradeResult, Question } from "./types";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init);
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json() as Promise<T>;
}

export function fetchQuestion(id: string): Promise<Question> {
  return request(`/api/questions/${id}`);
}

export function submitAnswers(id: string, answers: Answers): Promise<GradeResult> {
  return request(`/api/questions/${id}/answers`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ answers }),
  });
}
