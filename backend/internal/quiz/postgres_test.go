package quiz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// newPostgresStore は DATABASE_URL のデータベースに使い捨てのスキーマを作り、
// schema.sql と seed.sql を流した PostgresStore を返す。DATABASE_URL がなければスキップする。
func newPostgresStore(t *testing.T) *PostgresStore {
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
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
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

	for _, f := range []string{"../../db/schema.sql", "../../db/seed.sql"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, string(b)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	return NewPostgresStore(db)
}

func TestPostgresList(t *testing.T) {
	s := newPostgresStore(t)
	got, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []Summary{{ID: 2, Title: "図書館の貸出", BlankCount: 5}, {ID: 1, Title: "ECサイトの注文", BlankCount: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPostgresGetMatchesSample(t *testing.T) {
	s := newPostgresStore(t)
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want := SampleQuestions()[0]; !reflect.DeepEqual(got, want) {
		t.Fatalf("seed.sql の1問目と sample.go が一致しない\ngot  %+v\nwant %+v", got, want)
	}
}

func TestPostgresGetNotFound(t *testing.T) {
	s := newPostgresStore(t)
	if _, err := s.Get(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}
