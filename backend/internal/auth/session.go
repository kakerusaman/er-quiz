package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionTTL はログインしてからセッションが切れるまでの時間。
const SessionTTL = 30 * 24 * time.Hour

// ErrNoSession はセッションがない（期限切れ、ログアウト済みを含む）ときに返す。
var ErrNoSession = errors.New("no session")

// SessionStore はセッションの保存先。Redis 版（RedisSessionStore）と、メモリ版がある。
// key はトークンそのものではなく、そのハッシュ（sessionKey）を渡す。
type SessionStore interface {
	// Save は key にユーザーIDを ttl の間だけ保存する。
	Save(ctx context.Context, key string, userID int64, ttl time.Duration) error
	// UserID は key のユーザーIDを返す。なければ ErrNoSession。
	UserID(ctx context.Context, key string) (int64, error)
	// Delete は key のセッションを消す。なくてもエラーにしない。
	Delete(ctx context.Context, key string) error
}

// newToken は Cookie に入れるランダムなトークンを作る。
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("new token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// sessionKey はトークンから保存先のキーを作る。保存先にはハッシュだけを置き、
// 中身が漏れてもそのままログインに使えないようにする。
func sessionKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "session:" + hex.EncodeToString(sum[:])
}

// MemorySessionStore はメモリ上にセッションを持つ SessionStore。
type MemorySessionStore struct {
	mu       sync.Mutex
	sessions map[string]memorySession
	now      func() time.Time
}

type memorySession struct {
	userID  int64
	expires time.Time
}

// NewMemorySessionStore は空の MemorySessionStore を作る。
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: map[string]memorySession{}, now: time.Now}
}

func (s *MemorySessionStore) Save(_ context.Context, key string, userID int64, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[key] = memorySession{userID: userID, expires: s.now().Add(ttl)}
	return nil
}

func (s *MemorySessionStore) UserID(_ context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.sessions[key]
	if !ok || !s.now().Before(v.expires) {
		delete(s.sessions, key)
		return 0, ErrNoSession
	}
	return v.userID, nil
}

func (s *MemorySessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, key)
	return nil
}

// RedisSessionStore は Redis にセッションを保存する SessionStore。期限切れは Redis の TTL で消える。
type RedisSessionStore struct {
	rdb *redis.Client
}

// NewRedisSessionStore は接続済みの *redis.Client から RedisSessionStore を作る。
func NewRedisSessionStore(rdb *redis.Client) *RedisSessionStore { return &RedisSessionStore{rdb: rdb} }

func (s *RedisSessionStore) Save(ctx context.Context, key string, userID int64, ttl time.Duration) error {
	if err := s.rdb.Set(ctx, key, userID, ttl).Err(); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) UserID(ctx context.Context, key string) (int64, error) {
	v, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrNoSession
	}
	if err != nil {
		return 0, fmt.Errorf("get session: %w", err)
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("get session: %w", err)
	}
	return id, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
