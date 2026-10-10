// バックエンド（backend/internal/quiz/model.go）の JSON と対応する型。
export type Column = {
  id: string;
  name?: string;
  type: string;
  pk: boolean;
  fk: boolean;
  blankId?: string;
};

export type Table = {
  id: string;
  name?: string;
  blankId?: string;
  x: number;
  y: number;
  columns: Column[];
};

export type Relation = {
  id: string;
  from: string;
  to: string;
  cardinality: "1:1" | "1:N" | "N:M";
};

export type Blank = { id: string; choices: string[] };

export type Question = {
  id: string;
  title: string;
  tables: Table[];
  relations: Relation[];
  blanks: Blank[];
};

export type Answers = Record<string, string>;

export type GradeResult = {
  correct: Record<string, boolean>;
  score: number;
  total: number;
};
