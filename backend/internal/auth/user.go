// Package auth はユーザーの登録・ログインと、セッションによるログイン状態の管理を扱う。
package auth

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// User はログインできるユーザー。PasswordHash は API では返さない。
type User struct {
	ID           int64  `json:"id"`
	LoginID      string `json:"loginId"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

var (
	// ErrInvalidLoginID はログインIDの形式が決まりに合わないときに返す。
	ErrInvalidLoginID = errors.New("invalid login id")
	// ErrInvalidEmail はメールアドレスの形式が決まりに合わないときに返す。
	ErrInvalidEmail = errors.New("invalid email")
	// ErrInvalidPassword はパスワードの長さが決まりに合わないときに返す。
	ErrInvalidPassword = errors.New("invalid password")
	// ErrLoginIDTaken はログインIDがほかのユーザーに使われているときに返す。
	ErrLoginIDTaken = errors.New("login id taken")
	// ErrEmailTaken はメールアドレスがほかのユーザーに使われているときに返す。
	ErrEmailTaken = errors.New("email taken")
	// ErrUserNotFound はユーザーが見つからないときに返す。
	ErrUserNotFound = errors.New("user not found")
)

// loginIDPattern はログインIDに使える文字と長さ（小文字にそろえた後）。
var loginIDPattern = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

const (
	minPasswordLen = 8
	// maxPasswordBytes は bcrypt が扱える長さ。これより長い部分は無視されてしまうので受け付けない。
	maxPasswordBytes = 72
	maxEmailLen      = 254
)

// NormalizeLoginID はログインIDを小文字にそろえる。大文字と小文字は区別しない。
func NormalizeLoginID(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// NormalizeEmail はメールアドレスを小文字にそろえる。
func NormalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// ValidateSignup はそろえた後のログインID・メールアドレスと、パスワードを確かめる。
func ValidateSignup(loginID, email, password string) error {
	if !loginIDPattern.MatchString(loginID) {
		return ErrInvalidLoginID
	}
	if len(email) > maxEmailLen || !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}
	if utf8.RuneCountInString(password) < minPasswordLen || len(password) > maxPasswordBytes {
		return ErrInvalidPassword
	}
	return nil
}

// HashPassword はパスワードを bcrypt でハッシュにする。
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// dummyHash は存在しないログインIDでもパスワードを照合する時間をかけ、
// 応答時間からログインIDの有無を推測されないようにするためのハッシュ。
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy password"), bcrypt.DefaultCost)

// CheckPassword は hash が password のハッシュかを返す。hash が空なら dummyHash で照合して false を返す。
func CheckPassword(hash, password string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
