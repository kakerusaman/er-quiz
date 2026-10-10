package quiz

import "errors"

// ErrNotFound は問題が見つからないときに返す。
var ErrNotFound = errors.New("question not found")

// Store は問題の取得元。今はメモリ実装のみで、後で PostgreSQL 実装に差し替える。
type Store interface {
	List() []Summary
	Get(id string) (Question, error)
}

// MemoryStore はメモリ上に問題を持つ Store。
type MemoryStore struct {
	order     []string
	questions map[string]Question
}

// NewMemoryStore は与えた問題で MemoryStore を作る。
func NewMemoryStore(qs ...Question) *MemoryStore {
	s := &MemoryStore{questions: make(map[string]Question, len(qs))}
	for _, q := range qs {
		s.order = append(s.order, q.ID)
		s.questions[q.ID] = q
	}
	return s
}

func (s *MemoryStore) List() []Summary {
	out := make([]Summary, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, Summary{ID: id, Title: s.questions[id].Title})
	}
	return out
}

func (s *MemoryStore) Get(id string) (Question, error) {
	q, ok := s.questions[id]
	if !ok {
		return Question{}, ErrNotFound
	}
	return q, nil
}
