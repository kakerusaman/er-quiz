package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// newPostgresUserStore は DATABASE_URL のデータベースに使い捨てのスキーマを作り、
// schema.sql を流した PostgresUserStore を返す。DATABASE_URL がなければスキップする。
func newPostgresUserStore(t *testing.T) *PostgresUserStore {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL が未設定のため PostgreSQL のテストをスキップ")
	}
	ctx := context.Background()

	admin, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := fmt.Sprintf("test_auth_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") }) //nolint:errcheck

	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { db.Close() })

	b, err := os.ReadFile("../../db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(b)); err != nil {
		t.Fatalf("schema.sql: %v", err)
	}
	return NewPostgresUserStore(db)
}

// testUserStore は UserStore に共通の振る舞いを確かめる。
func testUserStore(t *testing.T, s UserStore) {
	t.Helper()
	ctx := context.Background()
	u, err := s.Create(ctx, User{LoginID: "taro", Email: "taro@example.com", PasswordHash: "hash"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Fatal("ID が入っていない")
	}
	if got, err := s.ByLoginID(ctx, "taro"); err != nil || got != u {
		t.Fatalf("ByLoginID = (%+v, %v), want %+v", got, err, u)
	}
	if got, err := s.ByID(ctx, u.ID); err != nil || got != u {
		t.Fatalf("ByID = (%+v, %v), want %+v", got, err, u)
	}
	if _, err := s.ByLoginID(ctx, "hanako"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("ないログインID: err = %v, want ErrUserNotFound", err)
	}
	if _, err := s.ByID(ctx, u.ID+100); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("ない ID: err = %v, want ErrUserNotFound", err)
	}
	if _, err := s.Create(ctx, User{LoginID: "taro", Email: "other@example.com", PasswordHash: "hash"}); !errors.Is(err, ErrLoginIDTaken) {
		t.Fatalf("ログインIDの重複: err = %v, want ErrLoginIDTaken", err)
	}
	if _, err := s.Create(ctx, User{LoginID: "hanako", Email: "taro@example.com", PasswordHash: "hash"}); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("メールアドレスの重複: err = %v, want ErrEmailTaken", err)
	}
}

func TestMemoryUserStore(t *testing.T) {
	testUserStore(t, NewMemoryUserStore())
}

func TestPostgresUserStore(t *testing.T) {
	testUserStore(t, newPostgresUserStore(t))
}
