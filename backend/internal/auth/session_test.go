package auth

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// testSessionStore は SessionStore に共通の振る舞いを確かめる。
func testSessionStore(t *testing.T, s SessionStore) {
	t.Helper()
	ctx := context.Background()
	key := sessionKey("token-" + time.Now().Format(time.RFC3339Nano))
	if _, err := s.UserID(ctx, key); !errors.Is(err, ErrNoSession) {
		t.Fatalf("保存前: err = %v, want ErrNoSession", err)
	}
	if err := s.Save(ctx, key, 5, time.Minute); err != nil {
		t.Fatal(err)
	}
	if id, err := s.UserID(ctx, key); err != nil || id != 5 {
		t.Fatalf("got (%d, %v), want (5, nil)", id, err)
	}
	if err := s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserID(ctx, key); !errors.Is(err, ErrNoSession) {
		t.Fatalf("削除後: err = %v, want ErrNoSession", err)
	}
	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("ないキーの削除: %v", err)
	}
}

func TestMemorySessionStore(t *testing.T) {
	testSessionStore(t, NewMemorySessionStore())
}

func TestMemorySessionStoreExpires(t *testing.T) {
	s := NewMemorySessionStore()
	now := time.Now()
	s.now = func() time.Time { return now }
	ctx := context.Background()
	if err := s.Save(ctx, "k", 1, time.Hour); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	if _, err := s.UserID(ctx, "k"); !errors.Is(err, ErrNoSession) {
		t.Fatalf("期限切れ: err = %v, want ErrNoSession", err)
	}
}

// newRedisSessionStore は REDIS_URL の Redis を使う。REDIS_URL がなければスキップする。
func newRedisSessionStore(t *testing.T) (*RedisSessionStore, *redis.Client) {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL が未設定のため Redis のテストをスキップ")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { rdb.Close() })
	return NewRedisSessionStore(rdb), rdb
}

func TestRedisSessionStore(t *testing.T) {
	s, _ := newRedisSessionStore(t)
	testSessionStore(t, s)
}

func TestRedisSessionStoreTTL(t *testing.T) {
	s, rdb := newRedisSessionStore(t)
	ctx := context.Background()
	key := sessionKey("ttl-" + time.Now().Format(time.RFC3339Nano))
	t.Cleanup(func() { rdb.Del(ctx, key) })
	if err := s.Save(ctx, key, 1, SessionTTL); err != nil {
		t.Fatal(err)
	}
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= SessionTTL-time.Minute || ttl > SessionTTL {
		t.Fatalf("TTL = %v, want about %v", ttl, SessionTTL)
	}
}
