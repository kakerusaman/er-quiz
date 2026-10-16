# ER図穴埋めクイズ

ER図の一部（テーブル名・カラム名など）が空欄になっていて、選択肢から埋めて答え合わせする学習用Webアプリです。

- フロントエンド: React + TypeScript + Vite、ER図の描画に React Flow（`@xyflow/react`）
- バックエンド: Go（ルーティングは Gin）、PostgreSQL（ドライバは `pgx`）、ログインのセッションは Redis（`go-redis`）

## 起動

```sh
# 最初に1回: PostgreSQL にテーブルとサンプル問題を作る（既存データは消える）
export DATABASE_URL="postgres://ユーザー名:パスワード@localhost:5432/データベース名?sslmode=disable"
make db-reset

# ログインのセッションを Redis に持つとき（なければサーバーのメモリに持ち、再起動でログアウトされる）
export REDIS_URL="redis://localhost:6379/0"

# ターミナル1: API（http://localhost:8080）。DATABASE_URL がなければメモリ上のサンプル問題で動く
make dev-backend

# ターミナル2: 画面（http://localhost:5173、/api はバックエンドへプロキシ）
cd frontend && npm install && cd ..
make dev-frontend
```

## テスト

```sh
make test
```

`DATABASE_URL` を設定していると、PostgreSQL の統合テストも動きます（使い捨てのスキーマを作って消すので、既存のテーブルには触れません）。`REDIS_URL` を設定していると、Redis の統合テストも動きます（`session:` で始まる使い捨てのキーだけを使います）。

HTTPS で公開するときは `SECURE_COOKIE=true` にして、ログインの Cookie に Secure を付けてください。

仕様は [docs/spec.md](docs/spec.md)、AIと開発するときの決まりは [CLAUDE.md](CLAUDE.md) にあります。
