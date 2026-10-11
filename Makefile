.PHONY: dev-backend dev-frontend db-reset test

dev-backend:
	cd backend && go run ./cmd/server

# DATABASE_URL のデータベースにテーブルを作り直し、サンプル問題を入れる（既存データは消える）
db-reset:
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -f backend/db/schema.sql -f backend/db/seed.sql

dev-frontend:
	cd frontend && npm run dev

test:
	cd backend && go vet ./... && go test ./...
	cd frontend && npm run typecheck && npm test
