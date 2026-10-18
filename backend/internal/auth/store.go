package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
)

// UserStore はユーザーの保存先。PostgreSQL 版（PostgresUserStore）と、メモリ版がある。
type UserStore interface {
	// Create はユーザーを登録し、ID を入れて返す。重複は ErrLoginIDTaken か ErrEmailTaken。
	Create(ctx context.Context, u User) (User, error)
	// ByLoginID はログインIDでユーザーを探す。見つからなければ ErrUserNotFound。
	ByLoginID(ctx context.Context, loginID string) (User, error)
	// ByID は ID でユーザーを探す。見つからなければ ErrUserNotFound。
	ByID(ctx context.Context, id int64) (User, error)
}

// MemoryUserStore はメモリ上にユーザーを持つ UserStore。
type MemoryUserStore struct {
	mu    sync.Mutex
	users []User
}

// NewMemoryUserStore は空の MemoryUserStore を作る。
func NewMemoryUserStore() *MemoryUserStore { return &MemoryUserStore{} }

func (s *MemoryUserStore) Create(_ context.Context, u User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.users {
		if x.LoginID == u.LoginID {
			return User{}, ErrLoginIDTaken
		}
		if x.Email == u.Email {
			return User{}, ErrEmailTaken
		}
	}
	u.ID = int64(len(s.users) + 1)
	s.users = append(s.users, u)
	return u, nil
}

func (s *MemoryUserStore) ByLoginID(_ context.Context, loginID string) (User, error) {
	return s.find(func(u User) bool { return u.LoginID == loginID })
}

func (s *MemoryUserStore) ByID(_ context.Context, id int64) (User, error) {
	return s.find(func(u User) bool { return u.ID == id })
}

func (s *MemoryUserStore) find(match func(User) bool) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if match(u) {
			return u, nil
		}
	}
	return User{}, ErrUserNotFound
}

// PostgresUserStore は t_users テーブルにユーザーを保存する UserStore。
type PostgresUserStore struct {
	db *sql.DB
}

// NewPostgresUserStore は接続済みの *sql.DB から PostgresUserStore を作る。
func NewPostgresUserStore(db *sql.DB) *PostgresUserStore { return &PostgresUserStore{db: db} }

func (s *PostgresUserStore) Create(ctx context.Context, u User) (User, error) {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO t_users (login_id, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		u.LoginID, u.Email, u.PasswordHash,
	).Scan(&u.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		switch pgErr.ConstraintName {
		case "t_users_login_id_key":
			return User{}, ErrLoginIDTaken
		case "t_users_email_key":
			return User{}, ErrEmailTaken
		}
	}
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *PostgresUserStore) ByLoginID(ctx context.Context, loginID string) (User, error) {
	return s.queryOne(ctx, `SELECT id, login_id, email, password_hash FROM t_users WHERE login_id = $1`, loginID)
}

func (s *PostgresUserStore) ByID(ctx context.Context, id int64) (User, error) {
	return s.queryOne(ctx, `SELECT id, login_id, email, password_hash FROM t_users WHERE id = $1`, id)
}

func (s *PostgresUserStore) queryOne(ctx context.Context, query string, arg any) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx, query, arg).Scan(&u.ID, &u.LoginID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
