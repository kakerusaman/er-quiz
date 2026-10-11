package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql に "pgx" ドライバを登録する

	"github.com/kakerusaman/er-quiz/backend/internal/quiz"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	h := quiz.NewHandler(newStore())
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}

// newStore は DATABASE_URL があれば PostgreSQL、なければメモリ上のサンプル問題を使う。
func newStore() quiz.Store {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Print("DATABASE_URL が未設定のため、メモリ上のサンプル問題で起動します")
		return quiz.NewMemoryStore(quiz.SampleQuestions()...)
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("PostgreSQL に接続できません: %v", err)
	}
	return quiz.NewPostgresStore(db)
}
