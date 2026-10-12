# データ設計

[仕様](../spec.md) をもとにした PostgreSQL のテーブル設計です。MVP の範囲を対象にしています。

## テーブル一覧

| テーブル | 内容 |
| --- | --- |
| questions | 問題（タイトル、業務の説明文、解説文） |
| entities | ER図のエンティティ（名前、図上の位置） |
| attributes | エンティティの属性（名前、主キー・外部キーか、表示順） |
| relationships | エンティティ間のリレーションシップ（両端の多重度） |
| choices | 解答群の選択肢 |
| blanks | 空欄（どの名前を隠すか、正解の選択肢） |
| attempts | 解答1回分（どの問題を、いつ、何点で解いたか） |
| attempt_answers | 解答の空欄ごとの中身（どの選択肢を置いて、正解だったか） |

## 各テーブルの列

### questions

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー（自動採番） |
| title | text | タイトル |
| scenario | text | 業務の説明文 |
| explanation | text | 解説文（問題ごとに1つ） |
| created_at | timestamptz | 作成日時 |

### entities

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| question_id | bigint | 問題（questions への外部キー） |
| name | text | エンティティ名。空欄になる場合も正解の名前を持つ（APIでは返さない） |
| x, y | integer | 図上の位置。配置は固定 |

### attributes

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| entity_id | bigint | エンティティ（entities への外部キー） |
| name | text | 属性名。空欄になる場合も正解の名前を持つ（APIでは返さない） |
| is_pk | boolean | 主キーか |
| is_fk | boolean | 外部キーか |
| position | integer | エンティティ内での表示順 |

データ型は持たない。午後問題のER図と同じく、名前と主キー・外部キーだけを表示する。

### relationships

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| question_id | bigint | 問題 |
| from_entity_id, to_entity_id | bigint | 線の両端のエンティティ |
| from_cardinality, to_cardinality | text | 両端の多重度 |

多重度は IE 記法の線の端に対応する次の4つです。

| 値 | 意味 | 線の端 |
| --- | --- | --- |
| `1` | ちょうど1 | ｜｜ |
| `0..1` | 0または1 | ○｜ |
| `1..N` | 1以上 | ｜と鳥の足 |
| `0..N` | 0以上 | ○と鳥の足 |

### choices

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| question_id | bigint | 問題 |
| label | text | 選択肢の文字列 |
| position | integer | 解答群での表示順 |

### blanks

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| question_id | bigint | 問題 |
| symbol | text | 空欄の記号（ア、イ、ウ…） |
| entity_id | bigint | エンティティ名を隠す場合のエンティティ。NULL可 |
| attribute_id | bigint | 属性名を隠す場合の属性。NULL可 |
| answer_choice_id | bigint | 正解の選択肢 |

- entity_id と attribute_id は、どちらか一方だけが入る。
- 1つの選択肢は1つの空欄にしか置けないため、同じ問題の中で answer_choice_id は重複させない（UNIQUE (question_id, answer_choice_id)）。

### attempts

| 列 | 型 | 内容 |
| --- | --- | --- |
| id | bigint | 主キー |
| question_id | bigint | 問題 |
| score | integer | 正解した空欄の数 |
| total | integer | 空欄の数 |
| created_at | timestamptz | 解答日時 |

ログインしていない人の解答も記録する。ログイン機能を作るときに user_id 列（未ログインは NULL）を追加する。

### attempt_answers

| 列 | 型 | 内容 |
| --- | --- | --- |
| attempt_id | bigint | 解答（attempts への外部キー） |
| blank_id | bigint | 空欄 |
| choice_id | bigint | 置いた選択肢。未回答は NULL |
| is_correct | boolean | 正解だったか |

主キーは (attempt_id, blank_id)。

## DDL

```sql
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

CREATE TABLE attempts (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question_id bigint      NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
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
```

## 補足

- 難易度（MVPの次）は attempt_answers から問題ごと・空欄ごとの正答率を集計して出す。
- 将来、主キー・外部キーやリレーションシップの線を空欄にするときは、blanks に参照先の列を足し、CHECK 制約を広げて対応する。
- 将来の関係スキーマ・SQL・正規化の問題は、questions に問題の種類を表す列を足して対応する。
