# CLAUDE.md

このリポジトリで作業するときの決まりです。

## 構成

- `backend/` Go のAPIサーバー。`cmd/server` が起動点、`internal/quiz` にモデル・保存・採点・HTTPハンドラ。
- `frontend/` React + TypeScript（Vite）。`src/components` に ER 図の部品、`src/api.ts` が API 呼び出し、`src/types.ts` が API の型。
- `docs/spec.md` 仕様。仕様を変えたらここも更新する。
- `docs/coding-rules.md` コーディングルール。コードを書く前に読み、これに従う。

## コマンド

- バックエンド: `cd backend && go vet ./... && go test ./...`
- フロント: `cd frontend && npm run typecheck && npm test && npm run build`
- 全部: `make test`

## 決まり

- 変更は1機能ずつ。最初に計画を出し、承認されてから実装する。
- 振る舞いを変えたらテストを足す（Go は `httptest`、フロントは Vitest + Testing Library）。
- 正解（`Blank.Answer`）は API レスポンスに含めない。採点は必ずサーバー側で行う。
- `backend/internal/quiz/model.go` の JSON と `frontend/src/types.ts` は手で揃えている。片方を変えたらもう片方も直す。
- React Flow のノードは `useNodesState` で持つ（毎回作り直すと計測済みサイズが消えてノードが表示されない）。
- Go は標準ライブラリ優先（例外: HTTP のルーティングとミドルウェアは Gin）。依存を足すときは理由を PR に書く。
- コミット前に上のコマンドがすべて通ることを確認する。
