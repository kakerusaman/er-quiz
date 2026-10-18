package quiz

import (
	"context"
	"errors"
	"sync"
)

// ErrNotFound は問題が見つからないときに返す。
var ErrNotFound = errors.New("question not found")

// Store は問題の取得元。PostgreSQL 版（PostgresStore）と、テスト用のメモリ版がある。
type Store interface {
	// List は問題一覧を新しい問題が上の順で返す。
	List(ctx context.Context) ([]Summary, error)
	// Get は正解と解説を含む問題全体を返す。見つからなければ ErrNotFound。
	Get(ctx context.Context, id int64) (QuestionData, error)
	// SaveAttempt は解答1回分（送られた回答と採点結果）を保存し、解答IDを返す。
	SaveAttempt(ctx context.Context, a Attempt) (int64, error)
}

// Attempt は保存する解答1回分。Answers は空欄ID → 置いた選択肢ID（未回答の空欄はない）。
// UserID は解いた人。未ログインなら nil。
type Attempt struct {
	QuestionID int64
	UserID     *int64
	Answers    map[int64]int64
	Result     Result
}

// MemoryStore はメモリ上に問題を持つ Store。後から渡した問題ほど新しいとみなす。
type MemoryStore struct {
	questions []QuestionData

	mu       sync.Mutex
	attempts []Attempt
}

// NewMemoryStore は与えた問題で MemoryStore を作る。
func NewMemoryStore(qs ...QuestionData) *MemoryStore {
	return &MemoryStore{questions: qs}
}

func (s *MemoryStore) List(context.Context) ([]Summary, error) {
	out := make([]Summary, 0, len(s.questions))
	for i := len(s.questions) - 1; i >= 0; i-- {
		out = append(out, s.questions[i].Summary())
	}
	return out, nil
}

func (s *MemoryStore) Get(_ context.Context, id int64) (QuestionData, error) {
	for _, q := range s.questions {
		if q.ID == id {
			return q, nil
		}
	}
	return QuestionData{}, ErrNotFound
}

func (s *MemoryStore) SaveAttempt(_ context.Context, a Attempt) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts = append(s.attempts, a)
	return int64(len(s.attempts)), nil
}

// Attempts は保存された解答を古い順に返す（テスト用）。
func (s *MemoryStore) Attempts() []Attempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Attempt(nil), s.attempts...)
}
