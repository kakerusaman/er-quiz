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

func TestPostgresSaveAttempt(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()
	q, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	answers := map[int64]int64{1: 1, 2: 2} // ウは未回答
	result := q.Grade(answers)
	id, err := s.SaveAttempt(ctx, Attempt{QuestionID: 1, Answers: answers, Result: result})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.SaveAttempt(ctx, Attempt{QuestionID: 1, Answers: answers, Result: result})
	if err != nil {
		t.Fatal(err)
	}
	if id2 == id {
		t.Fatalf("attempt ids are equal: %d", id)
	}

	var questionID int64
	var score, total int
	if err := s.db.QueryRowContext(ctx,
		`SELECT question_id, score, total FROM attempts WHERE id = $1`, id,
	).Scan(&questionID, &score, &total); err != nil {
		t.Fatal(err)
	}
	if questionID != 1 || score != 1 || total != 3 {
		t.Fatalf("attempt = (%d, %d, %d)", questionID, score, total)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT blank_id, choice_id, is_correct FROM attempt_answers WHERE attempt_id = $1 ORDER BY blank_id`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var blank int64
		var choice sql.NullInt64
		var correct bool
		if err := rows.Scan(&blank, &choice, &correct); err != nil {
			t.Fatal(err)
		}
		got = append(got, fmt.Sprintf("%d:%v:%v", blank, choice, correct))
	}
	want := []string{"1:{1 true}:true", "2:{2 true}:false", "3:{0 false}:false"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("attempt_answers = %v, want %v", got, want)
	}
}

func TestPostgresSaveAttemptUser(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()
	var userID int64
	if err := s.db.QueryRowContext(ctx,
		`INSERT INTO t_users (login_id, email, password_hash) VALUES ('taro', 'taro@example.com', 'x') RETURNING id`,
	).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	q, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	result := q.Grade(map[int64]int64{})
	withUser, err := s.SaveAttempt(ctx, Attempt{QuestionID: 1, UserID: &userID, Answers: map[int64]int64{}, Result: result})
	if err != nil {
		t.Fatal(err)
	}
	anonymous, err := s.SaveAttempt(ctx, Attempt{QuestionID: 1, Answers: map[int64]int64{}, Result: result})
	if err != nil {
		t.Fatal(err)
	}

	for id, want := range map[int64]sql.NullInt64{withUser: {Int64: userID, Valid: true}, anonymous: {}} {
		var got sql.NullInt64
		if err := s.db.QueryRowContext(ctx, `SELECT user_id FROM attempts WHERE id = $1`, id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("attempt %d: user_id = %v, want %v", id, got, want)
		}
	}
}
