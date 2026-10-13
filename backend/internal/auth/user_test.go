package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSignup(t *testing.T) {
	tests := []struct {
		name, loginID, email, password string
		want                           error
	}{
		{"正しい入力", "taro_01", "taro@example.com", "password", nil},
		{"ログインIDが短い", "ab", "taro@example.com", "password", ErrInvalidLoginID},
		{"ログインIDが長い", strings.Repeat("a", 21), "taro@example.com", "password", ErrInvalidLoginID},
		{"ログインIDに使えない文字", "taro-01", "taro@example.com", "password", ErrInvalidLoginID},
		{"メールアドレスに@がない", "taro", "taro.example.com", "password", ErrInvalidEmail},
		{"メールアドレスが長い", "taro", strings.Repeat("a", 250) + "@a.jp", "password", ErrInvalidEmail},
		{"パスワードが短い", "taro", "taro@example.com", "1234567", ErrInvalidPassword},
		{"パスワードが72バイトを超える", "taro", "taro@example.com", strings.Repeat("a", 73), ErrInvalidPassword},
		{"日本語のパスワードは文字数で数える", "taro", "taro@example.com", "あいうえおかきく", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateSignup(tt.loginID, tt.email, tt.password); !errors.Is(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	if got := NormalizeLoginID(" Taro_01 "); got != "taro_01" {
		t.Fatalf("NormalizeLoginID = %q", got)
	}
	if got := NormalizeEmail("Taro@Example.COM"); got != "taro@example.com" {
		t.Fatalf("NormalizeEmail = %q", got)
	}
}

func TestPassword(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "password") {
		t.Fatal("正しいパスワードが通らない")
	}
	if CheckPassword(hash, "passw0rd") {
		t.Fatal("違うパスワードが通る")
	}
	if CheckPassword("", "dummy password") {
		t.Fatal("ハッシュがない（ユーザーがいない）のに通る")
	}
}
