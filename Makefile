.PHONY: dev-backend dev-frontend test

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

test:
	cd backend && go vet ./... && go test ./...
	cd frontend && npm run typecheck && npm test
