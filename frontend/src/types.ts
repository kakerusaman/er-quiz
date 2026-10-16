// バックエンド（backend/internal/quiz/model.go、backend/internal/auth/user.go）の JSON と対応する型。

export type BlankRef = { id: number; symbol: string };

export type Attribute = {
  id: number;
  /** 空欄のときは null */
  name: string | null;
  isPk: boolean;
  isFk: boolean;
  blank: BlankRef | null;
};

export type Entity = {
  id: number;
  /** 空欄のときは null */
  name: string | null;
  x: number;
  y: number;
  blank: BlankRef | null;
  attributes: Attribute[];
};

export type Cardinality = "1" | "0..1" | "1..N" | "0..N";

export type Relationship = {
  id: number;
  fromEntityId: number;
  toEntityId: number;
  fromCardinality: Cardinality;
  toCardinality: Cardinality;
};

export type Choice = { id: number; label: string };

export type Question = {
  id: number;
  title: string;
  scenario: string;
  entities: Entity[];
  relationships: Relationship[];
  choices: Choice[];
};

export type Summary = { id: number; title: string; blankCount: number };

/** 空欄ID → 選択肢ID */
export type Answers = Record<number, number>;

export type BlankResult = {
  blankId: number;
  correct: boolean;
  /** 正解の選択肢ID */
  answerChoiceId: number;
};

export type AttemptResult = {
  attemptId: number;
  score: number;
  total: number;
  /** 空欄ID順 */
  results: BlankResult[];
  explanation: string;
};

/** ログインしているユーザー */
export type User = { id: number; loginId: string; email: string };
