# ER図穴埋めクイズ

ER図の一部（テーブル名・カラム名など）が空欄になっていて、選択肢から埋めて答え合わせする学習用Webアプリです。

- フロントエンド: React + TypeScript + Vite、ER図の描画に React Flow（`@xyflow/react`）
- バックエンド: Go（標準ライブラリの `net/http`）

## 起動

```sh
# ターミナル1: API（http://localhost:8080）
make dev-backend

# ターミナル2: 画面（http://localhost:5173、/api はバックエンドへプロキシ）
cd frontend && npm install && cd ..
make dev-frontend
```

## テスト

```sh
make test
```

仕様は [docs/spec.md](docs/spec.md)、AIと開発するときの決まりは [CLAUDE.md](CLAUDE.md) にあります。
