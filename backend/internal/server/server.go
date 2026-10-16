// Package server は API のルートとミドルウェアをまとめて、1つの http.Handler にする。
package server

import (
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kakerusaman/er-quiz/backend/internal/auth"
	"github.com/kakerusaman/er-quiz/backend/internal/quiz"
)

// Config はサーバーが使う保存先と設定。
type Config struct {
	Questions quiz.Store
	Users     auth.UserStore
	Sessions  auth.SessionStore
	// SecureCookie が true なら Cookie に Secure を付ける（HTTPS で動かす本番用）。
	SecureCookie bool
}

// New は /api 以下のルートを持つ http.Handler を返す。
func New(cfg Config) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// メソッド違いは 404 ではなく 405 を返す（net/http の ServeMux と同じ振る舞い）。
	r.HandleMethodNotAllowed = true

	ah := &auth.Handler{Users: cfg.Users, Sessions: cfg.Sessions, SecureCookie: cfg.SecureCookie}
	api := r.Group("/api", requireJSON, ah.LoadUser)
	ah.Register(api)
	quiz.Register(api, cfg.Questions)
	return r
}

// requireJSON は JSON 以外の POST を 415 で断る。フォームの送信で
// 他サイトから POST させる CSRF を、SameSite=Lax の Cookie と合わせて防ぐ。
func requireJSON(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		mt, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || mt != "application/json" {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "content type must be application/json"})
			return
		}
	}
	c.Next()
}
