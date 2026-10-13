package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CookieName はセッションのトークンを入れる Cookie の名前。
const CookieName = "session"

// userKey は LoadUser が gin.Context にログイン中のユーザーを入れるときのキー。
const userKey = "auth.user"

// Handler は登録・ログイン・ログアウトの API と、ログイン状態を読むミドルウェアを持つ。
type Handler struct {
	Users    UserStore
	Sessions SessionStore
	// SecureCookie が true なら Cookie に Secure を付ける（HTTPS で動かす本番用）。
	SecureCookie bool
}

// Register は api グループに /signup、/login、/logout、/me を登録する。
func (h *Handler) Register(api *gin.RouterGroup) {
	api.POST("/signup", h.signup)
	api.POST("/login", h.login)
	api.POST("/logout", h.logout)
	api.GET("/me", RequireLogin, func(c *gin.Context) {
		u, _ := CurrentUser(c)
		c.JSON(http.StatusOK, u)
	})
}

// LoadUser はセッションの Cookie があればユーザーを読み込み、CurrentUser で取れるようにする。
// Cookie がない・セッションが切れているときは、未ログインとして次へ進む。
func (h *Handler) LoadUser(c *gin.Context) {
	token, err := c.Cookie(CookieName)
	if err != nil || token == "" {
		c.Next()
		return
	}
	ctx := c.Request.Context()
	id, err := h.Sessions.UserID(ctx, sessionKey(token))
	if errors.Is(err, ErrNoSession) {
		c.Next()
		return
	}
	if err != nil {
		abortInternal(c, err)
		return
	}
	u, err := h.Users.ByID(ctx, id)
	if errors.Is(err, ErrUserNotFound) {
		c.Next()
		return
	}
	if err != nil {
		abortInternal(c, err)
		return
	}
	c.Set(userKey, u)
	c.Next()
}

// RequireLogin は未ログインなら 401 を返して止めるミドルウェア。LoadUser の後に付ける。
func RequireLogin(c *gin.Context) {
	if _, ok := CurrentUser(c); !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}
	c.Next()
}

// CurrentUser はログイン中のユーザーを返す。未ログインなら false。
func CurrentUser(c *gin.Context) (User, bool) {
	v, ok := c.Get(userKey)
	if !ok {
		return User{}, false
	}
	u, ok := v.(User)
	return u, ok
}

type credentials struct {
	LoginID  string `json:"loginId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) signup(c *gin.Context) {
	var in credentials
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	loginID, email := NormalizeLoginID(in.LoginID), NormalizeEmail(in.Email)
	if err := ValidateSignup(loginID, email, in.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		abortInternal(c, err)
		return
	}
	u, err := h.Users.Create(c.Request.Context(), User{LoginID: loginID, Email: email, PasswordHash: hash})
	if errors.Is(err, ErrLoginIDTaken) || errors.Is(err, ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		abortInternal(c, err)
		return
	}
	if !h.startSession(c, u) {
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *Handler) login(c *gin.Context) {
	var in credentials
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	u, err := h.Users.ByLoginID(c.Request.Context(), NormalizeLoginID(in.LoginID))
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		abortInternal(c, err)
		return
	}
	// ログインIDがないときも照合の時間をかけ、ログインID違いかパスワード違いかを区別させない。
	if !CheckPassword(u.PasswordHash, in.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !h.startSession(c, u) {
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) logout(c *gin.Context) {
	if !h.endSession(c) {
		return
	}
	h.setCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

// startSession は今のセッションがあれば消し、u の新しいセッションを作って Cookie に入れる。
// 失敗したらエラーを書いて false を返す。
func (h *Handler) startSession(c *gin.Context, u User) bool {
	if !h.endSession(c) {
		return false
	}
	token, err := newToken()
	if err != nil {
		abortInternal(c, err)
		return false
	}
	if err := h.Sessions.Save(c.Request.Context(), sessionKey(token), u.ID, SessionTTL); err != nil {
		abortInternal(c, err)
		return false
	}
	h.setCookie(c, token, int(SessionTTL.Seconds()))
	return true
}

// endSession は Cookie のセッションがあれば消す。失敗したらエラーを書いて false を返す。
func (h *Handler) endSession(c *gin.Context) bool {
	token, err := c.Cookie(CookieName)
	if err != nil || token == "" {
		return true
	}
	if err := h.Sessions.Delete(c.Request.Context(), sessionKey(token)); err != nil {
		abortInternal(c, err)
		return false
	}
	return true
}

// setCookie はセッションの Cookie を書く。maxAge が負なら Cookie を消す。
func (h *Handler) setCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func abortInternal(c *gin.Context, err error) {
	log.Printf("internal error: %v", err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}
