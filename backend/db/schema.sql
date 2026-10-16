-- PostgreSQL のテーブル定義。docs/design/data.md の DDL と同じ内容。
-- 開発用: 実行すると既存のテーブルとデータを消して作り直す。

DROP TABLE IF EXISTS attempt_answers, attempts, t_users, blanks, choices, relationships, attributes, entities, questions CASCADE;

CREATE TABLE questions (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       text        NOT NULL,
    scenario    text        NOT NULL,
    explanation text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE entities (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id bigint  NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    name        text    NOT NULL,
    x           integer NOT NULL,
    y           integer NOT NULL,
    UNIQUE (question_id, id)
);

CREATE TABLE attributes (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    entity_id bigint  NOT NULL REFERENCES entities (id) ON DELETE CASCADE,
    name      text    NOT NULL,
    is_pk     boolean NOT NULL DEFAULT false,
    is_fk     boolean NOT NULL DEFAULT false,
    position  integer NOT NULL,
    UNIQUE (entity_id, position)
);

CREATE TABLE relationships (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id      bigint NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    from_entity_id   bigint NOT NULL,
    to_entity_id     bigint NOT NULL,
    from_cardinality text   NOT NULL CHECK (from_cardinality IN ('1', '0..1', '1..N', '0..N')),
    to_cardinality   text   NOT NULL CHECK (to_cardinality IN ('1', '0..1', '1..N', '0..N')),
    -- 線の両端が同じ問題のエンティティであることを保証する
    FOREIGN KEY (question_id, from_entity_id) REFERENCES entities (question_id, id) ON DELETE CASCADE,
    FOREIGN KEY (question_id, to_entity_id)   REFERENCES entities (question_id, id) ON DELETE CASCADE
);

CREATE TABLE choices (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id bigint  NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    label       text    NOT NULL,
    position    integer NOT NULL,
    UNIQUE (question_id, id),
    UNIQUE (question_id, label),
    UNIQUE (question_id, position)
);

CREATE TABLE blanks (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id      bigint NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    symbol           text   NOT NULL,
    entity_id        bigint UNIQUE REFERENCES entities (id) ON DELETE CASCADE,
    attribute_id     bigint UNIQUE REFERENCES attributes (id) ON DELETE CASCADE,
    answer_choice_id bigint NOT NULL,
    UNIQUE (question_id, symbol),
    -- 選択肢は1つの空欄にしか置けないので、正解の選択肢も問題の中で重複させない
    UNIQUE (question_id, answer_choice_id),
    -- 正解の選択肢が同じ問題の解答群にあることを保証する
    FOREIGN KEY (question_id, answer_choice_id) REFERENCES choices (question_id, id),
    -- エンティティ名か属性名の、どちらか一方だけを隠す
    CHECK (num_nonnulls(entity_id, attribute_id) = 1)
);

-- ログインID・メールアドレスは小文字にそろえて保存する（大文字と小文字を区別しない）。
-- 重複のエラーを見分けるため、UNIQUE 制約の名前（t_users_login_id_key、t_users_email_key）を Go のコードで使っている。
CREATE TABLE t_users (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login_id      text        NOT NULL UNIQUE CHECK (login_id ~ '^[a-z0-9_]{3,20}$'),
    email         text        NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE attempts (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id bigint      NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    -- 解いた人。未ログインなら NULL。
    user_id     bigint      REFERENCES t_users (id) ON DELETE SET NULL,
    score       integer     NOT NULL CHECK (score >= 0),
    total       integer     NOT NULL CHECK (total >= score),
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE attempt_answers (
    attempt_id bigint  NOT NULL REFERENCES attempts (id) ON DELETE CASCADE,
    blank_id   bigint  NOT NULL REFERENCES blanks (id) ON DELETE CASCADE,
    choice_id  bigint  REFERENCES choices (id) ON DELETE CASCADE,
    is_correct boolean NOT NULL,
    PRIMARY KEY (attempt_id, blank_id)
);
