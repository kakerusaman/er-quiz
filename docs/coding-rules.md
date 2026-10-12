# コーディングルール

このリポジトリのコードを書くときの決まりです。今のコードの書き方をもとにしています。作業の進め方（計画→承認→実装、コミット前のコマンドなど）は [CLAUDE.md](../CLAUDE.md) を見てください。

迷ったら、まわりのコードに合わせます。このファイルと今のコードが食い違っていたら、どちらかを直して揃えます。

## 共通

- コメント、テスト名、エラー画面の文言、ドキュメントは日本語で書く。識別子（変数名・関数名・型名）は英語。
- コメントは「なぜそうするか」や、コードから読み取れない前提を書く。コードをそのまま言い換えるだけのコメントは書かない。
- 仕様・設計に関わる名前は [docs/spec.md](spec.md) と [docs/design/](design/) の用語に合わせる（問題 = question、空欄 = blank、解答群の選択肢 = choice、解答 = attempt、業務の説明文 = scenario など）。
- 1つの変更で1つの機能。ついでのリファクタリングや整形は別の変更にする。
- 正解（`Answers`、`Blank.Answer`）と解説は、答え合わせのレスポンス以外で API から返さない。採点は必ずサーバー側で行う。

## Go（`backend/`）

### 書き方

- 整形は `gofmt`、チェックは `go vet ./...`。どちらも通らないコードはコミットしない。
- 標準ライブラリを優先する。HTTP は `net/http` の `ServeMux`（`"GET /api/questions/{id}"` のようなパターン）、DB は `database/sql` + `pgx` ドライバ。依存を足すときは理由を PR に書く。
- パッケージは機能単位でまとめる（今は `internal/quiz` に、モデル・保存・採点・ハンドラ）。外から使わせないものは `internal/` に置く。
- 公開する型・関数・変数には、名前で始まる日本語のドキュメントコメントを付ける（例: `// Grade は空欄ID → 選択肢ID の回答を採点する。`）。パッケージには `// Package quiz は…` を書く。
- コンストラクタは `NewXxx`。インターフェースは使う側に置き、小さく保つ（例: `Store`）。

### エラー

- 呼び出し元が判定するエラーは `ErrXxx` のパッケージ変数にし、`errors.Is` で判定する（`ErrNotFound`、`ErrInvalidAnswer`）。
- 詳しい情報を足すときは `fmt.Errorf("get question %d: %w", id, err)` のように `%w` で包む。メッセージは英語の小文字始まり、何をしていたかを先頭に書く。
- ハンドラでは内部エラーの詳細をレスポンスに出さない。ログに書いて、クライアントには `{"error": "internal error"}` を返す（`writeError` を使う）。
- `log.Fatal` は `main` の起動時だけ。

### HTTP・JSON

- レスポンスは `writeJSON` / `writeError` を通して返す。エラーは `{"error": "..."}` の形。
- JSON のキーは camelCase。構造体には必ず `json:"..."` タグを付ける。
- 正解を含む `QuestionData` をそのまま返さない。問題を返すときは必ず `Public()` を通す。
- `model.go` の JSON を変えたら、`frontend/src/types.ts` と `docs/design/api.md` も同じ変更で直す。

### DB

- SQL は文字列で書き、値は必ずプレースホルダ（`$1`）で渡す。値を文字列連結で SQL に入れない。
- `QueryContext` などの `Context` 付きのメソッドを使い、リクエストの `ctx` を渡す。
- 複数のテーブルをまとめて読む・書くときはトランザクションを使う。
- テーブル名・列名の決まりは「テーブル設計（PostgreSQL）」を見る。

### テスト

- テストはテスト対象と同じパッケージに `xxx_test.go` で置く。
- HTTP の振る舞いは `httptest.NewServer(NewHandler(NewMemoryStore(...)))` で、API を外から呼んでテストする。
- PostgreSQL のテストは `DATABASE_URL` があるときだけ動かし、なければ `t.Skip` する。使い捨てのスキーマを作って `t.Cleanup` で消す。
- ヘルパーには `t.Helper()` を付ける。後片付けは `t.Cleanup` で書く。
- 失敗メッセージは `got ..., want ...` の形にし、意図が分かりにくいときは日本語で補う（例: `want %+v（新しい問題が上）`）。
- 正解が漏れていないことなど、守るべき約束はテストで確かめる（例: `TestGetQuestionHidesAnswers`）。

## テーブル設計（PostgreSQL）

### テーブル名

- テーブル名は、マスタなら `m_`、トランザクションなら `t_` で始める。
  - マスタ: 前もって登録しておき、アプリを使っても増えたり変わったりしないデータ。問題とその中身（ER図、選択肢、空欄と正解）。
  - トランザクション: 利用者の操作のたびに増えていくデータ。解答の記録。
  - どちらか迷うときは「解く人が操作したときに行が増えるか」で決め、PR に理由を書く。
- 接頭辞のあとは英語の snake_case の複数形にする（例: `m_questions`、`t_attempt_answers`）。
- 今のテーブルとの対応は次のとおり。

| 区分 | 今の名前 | ルールに沿った名前 |
| --- | --- | --- |
| マスタ | questions | m_questions |
| マスタ | entities | m_entities |
| マスタ | attributes | m_attributes |
| マスタ | relationships | m_relationships |
| マスタ | choices | m_choices |
| マスタ | blanks | m_blanks |
| トランザクション | attempts | t_attempts |
| トランザクション | attempt_answers | t_attempt_answers |

### 列・制約

- 列名は英語の snake_case。接頭辞は付けない。
- 主キーは `id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY`。中間テーブルなど、行を特定するのに複数の列が要るときは複合主キーにしてよい（例: `t_attempt_answers` の `(attempt_id, blank_id)`）。
- 外部キーの列は、参照先テーブルの単数形 + `_id`（接頭辞は除く）にする（例: `m_questions` を参照するなら `question_id`）。同じテーブルを役割違いで参照するときは、前に役割を付ける（`from_entity_id`、`answer_choice_id`）。
- 真偽値の列は `is_` で始める（`is_pk`、`is_correct`）。表示順は `position`、作成日時は `created_at timestamptz NOT NULL DEFAULT now()`。
- 列は原則 `NOT NULL`。値がないことに意味があるときだけ NULL を許し、その意味をコメントか `docs/design/data.md` に書く。
- 守るべき整合性は、アプリだけでなく制約（外部キー、`UNIQUE`、`CHECK`）でも守る。制約の意図が名前から読み取れないときは、DDL にコメントを書く。
- テーブルや列を変えたら、`backend/db/schema.sql`、`docs/design/data.md`（テーブル一覧・列・DDL）、必要なら `seed.sql` と Go の SQL を同じ変更で直す。

## TypeScript / React（`frontend/`）

### 書き方

- `tsconfig.json` の `strict`、`noUnusedLocals`、`noUnusedParameters` を緩めない。`any` は使わない（外から来た値は型を付けて受けるか `unknown` から絞る）。
- 整形用のツールは入れていない。今のコードに合わせる: インデント2スペース、ダブルクォート、セミコロンあり、複数行の末尾カンマあり、1行はおおむね120文字まで。
- 型は `type` で書く（`interface` は使わない）。型だけの import は `import type` / `type Xxx` を使う。
- export は名前付き（`export function Xxx`）にし、`export default` は使わない。
- 値がないことは `undefined`（省略可能なプロパティ、`useState<T>()`）で表す。`null` は API の JSON で `null` が来るところ（`name`、`blank`）だけで使う。
- 関数・変数は camelCase、コンポーネントと型は PascalCase、フックは `useXxx`。ファイル名はコンポーネントなら `ErDiagram.tsx` のように PascalCase、それ以外は `answers.ts` のように camelCase。

### ファイルの置き場所

- `src/pages/` 画面単位のコンポーネント（`QuestionList`、`QuestionPage`）。データの取得と画面の状態はここで持つ。
- `src/components/` 画面の部品。表示と、props で受けた値・コールバックを使う操作だけにする。
- `src/api.ts` API 呼び出し。`fetch` はここ以外で直接使わない。
- `src/types.ts` API の型。`backend/internal/quiz/model.go` の JSON と手で揃える。
- ロジック（回答の更新、当たり判定など）は React に依存しない関数として `src/xxx.ts` に切り出し、単体テストを書く（`answers.ts`、`collision.ts`）。

### コンポーネント

- 関数コンポーネントだけを使う。props の型はファイル内で `type Props = {...}` と書く。
- コンポーネントの前に、何を表示するものかを1〜2行のコメントで書く。
- 非同期の取得は `useEffect` で行い、アンマウント後や `id` が変わった後に古い結果で state を更新しないようにする（`let active = true` のパターン）。
- 依存を足さずに書けるものは自前で書く（ルーティングは `router.ts` の History API、画面幅は `useIsNarrow`）。
- React Flow のノードは `useNodesState` で持ち、回答や結果が変わったときは `setNodes` で `data` だけ差し替える（ノードを作り直すと計測済みのサイズが消え、ノードが表示されなくなる）。
- React Flow の図の中でドラッグさせる要素には `nodrag nopan` クラスを付ける。

### アクセシビリティ

- ボタンは `<button>`、まとまりは `<section>` にして見出しを `aria-labelledby` で結ぶ。記号だけのボタンやチップには `aria-label` を付ける（例: `aria-label="拡大"`、`aria-label={`空欄 ${symbol}`}`）。
- 飾りのアイコンには `aria-hidden="true"` を付ける。
- テストはロールと名前（`getByRole("button", { name: "答え合わせ" })`）で要素を探すので、画面に出る名前を変えたらテストも直す。

### スタイル

- CSS は `src/index.css` 1つにまとめる（React Flow の CSS の読み込みは除く）。CSS Modules や CSS-in-JS は使わない。
- 色は `:root` の CSS 変数（`--primary`、`--correct` など）を使い、値を直接書かない。新しい色は変数を足し、`docs/design/screen-detail.md` の色の表も直す。
- クラス名は kebab-case（`blank-chip`、`question-list`）。状態は `correct`、`filled`、`over` のような追加のクラスで表す。
- インラインの `style` は使わない。どうしても値が動的に決まるとき（座標など）だけにする。

### テスト

- Vitest + Testing Library。テストはテスト対象の隣に `xxx.test.ts(x)` で置く。
- `describe` は関数・コンポーネント名、`it` は振る舞いを日本語で書く（例: `it("空欄の外へ出すと空に戻る", ...)`）。
- 画面のテストは `fetch` を `vi.fn` で差し替え、`App` を描画してユーザーの操作（`userEvent`）で確かめる。実装の内部（state や関数の呼び出し回数）ではなく、画面に出るものを確かめる。

## Git・PR

- コミットメッセージは日本語で、何ができるようになったか・何を変えたかを1行で書く（例: `同じ選択肢を1つの空欄にしか置けないようにする`）。ドキュメントだけの変更は `docs:` を付ける。
- PR は1機能ずつ。仕様を変えたら `docs/spec.md` を、設計を変えたら `docs/design/` を同じ PR で直す。
- コミット前に `make test`（フロントは `npm run build` も）が通ることを確認する。CI でも同じものを動かしている。
