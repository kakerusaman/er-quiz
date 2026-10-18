package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql に "pgx" ドライバを登録する
	"github.com/redis/go-redis/v9"

	"github.com/kakerusaman/er-quiz/backend/internal/auth"
	"github.com/kakerusaman/er-quiz/backend/internal/quiz"
	"github.com/kakerusaman/er-quiz/backend/internal/server"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	cfg := server.Config{
		Sessions:     newSessionStore(),
		SecureCookie: os.Getenv("SECURE_COOKIE") == "true",
	}
	cfg.Questions, cfg.Users = newStores()
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.New(cfg)))
}

// newStores は DATABASE_URL があれば PostgreSQL、なければメモリ上のサンプル問題と空のユーザーを使う。
func newStores() (quiz.Store, auth.UserStore) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Print("DATABASE_URL が未設定のため、メモリ上のサンプル問題で起動します")
		return quiz.NewMemoryStore(quiz.SampleQuestions()...), auth.NewMemoryUserStore()
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
	return quiz.NewPostgresStore(db), auth.NewPostgresUserStore(db)
}

// newSessionStore は REDIS_URL があれば Redis、なければメモリにセッションを持つ。
func newSessionStore() auth.SessionStore {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		log.Print("REDIS_URL が未設定のため、セッションをメモリに持ちます（再起動するとログアウトされます）")
		return auth.NewMemorySessionStore()
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis に接続できません: %v", err)
	}
	return auth.NewRedisSessionStore(rdb)
}
