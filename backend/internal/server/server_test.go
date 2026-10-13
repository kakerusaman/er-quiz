package server

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kakerusaman/er-quiz/backend/internal/auth"
	"github.com/kakerusaman/er-quiz/backend/internal/quiz"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

type testServer struct {
	url       string
	questions *quiz.MemoryStore
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{
		questions: quiz.NewMemoryStore(quiz.SampleQuestions()...),
	}
	srv := httptest.NewServer(New(Config{Questions: ts.questions, Users: auth.NewMemoryUserStore(), Sessions: auth.NewMemorySessionStore()}))
	t.Cleanup(srv.Close)
	ts.url = srv.URL
	return ts
}

// newClient は Cookie を覚えるクライアント（ブラウザの代わり）を作る。
func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

// call は JSON の body を送り、ステータスを返す。v があればレスポンスを読む。
func call(t *testing.T, c *http.Client, method, url, body string, v any) int {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if v != nil {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			t.Fatal(err)
		}
	}
	return res.StatusCode
}

const signupBody = `{"loginId":"Taro_01","email":"Taro@Example.com","password":"password"}`

func TestSignupLoginLogout(t *testing.T) {
	ts := newTestServer(t)
	c := newClient(t)

	if status := call(t, c, "GET", ts.url+"/api/me", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("登録前の /me: status = %d", status)
	}

	var u map[string]any
	if status := call(t, c, "POST", ts.url+"/api/signup", signupBody, &u); status != http.StatusCreated {
		t.Fatalf("signup: status = %d", status)
	}
	if u["loginId"] != "taro_01" || u["email"] != "taro@example.com" {
		t.Fatalf("signup = %v（小文字にそろえる）", u)
	}
	if _, ok := u["passwordHash"]; ok {
		t.Fatalf("password hash leaked: %v", u)
	}

	// 登録するとそのままログインしている。
	var me auth.User
	if status := call(t, c, "GET", ts.url+"/api/me", "", &me); status != http.StatusOK || me.LoginID != "taro_01" {
		t.Fatalf("登録後の /me: status = %d, user = %+v", status, me)
	}

	if status := call(t, c, "POST", ts.url+"/api/logout", "{}", nil); status != http.StatusNoContent {
		t.Fatalf("logout: status = %d", status)
	}
	if status := call(t, c, "GET", ts.url+"/api/me", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("ログアウト後の /me: status = %d", status)
	}

	// 大文字のログインIDでもログインできる。
	if status := call(t, c, "POST", ts.url+"/api/login", `{"loginId":"TARO_01","password":"password"}`, &me); status != http.StatusOK {
		t.Fatalf("login: status = %d", status)
	}
	if status := call(t, c, "GET", ts.url+"/api/me", "", nil); status != http.StatusOK {
		t.Fatalf("ログイン後の /me: status = %d", status)
	}
}

func TestLogoutDeletesSession(t *testing.T) {
	ts := newTestServer(t)
	c := newClient(t)
	call(t, c, "POST", ts.url+"/api/signup", signupBody, nil)
	cookies := c.Jar.Cookies(mustParse(t, ts.url))
	call(t, c, "POST", ts.url+"/api/logout", "{}", nil)

	// ログアウト前の Cookie を持ち続けても、ログイン中には戻らない。
	old := newClient(t)
	old.Jar.SetCookies(mustParse(t, ts.url), cookies)
	if status := call(t, old, "GET", ts.url+"/api/me", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("古い Cookie の /me: status = %d", status)
	}
}

func TestSessionCookie(t *testing.T) {
	ts := newTestServer(t)
	req, _ := http.NewRequest("POST", ts.url+"/api/signup", strings.NewReader(signupBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v", cookies)
	}
	ck := cookies[0]
	if ck.Name != auth.CookieName || !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/" || ck.MaxAge != int(auth.SessionTTL.Seconds()) {
		t.Fatalf("cookie = %+v", ck)
	}
}

func TestSignupErrors(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
		error      string
	}{
		{"ログインIDが短い", `{"loginId":"ab","email":"a@example.com","password":"password"}`, http.StatusBadRequest, "invalid login id"},
		{"メールアドレスの形でない", `{"loginId":"hanako","email":"hanako","password":"password"}`, http.StatusBadRequest, "invalid email"},
		{"パスワードが短い", `{"loginId":"hanako","email":"a@example.com","password":"short"}`, http.StatusBadRequest, "invalid password"},
		{"壊れたJSON", `{"loginId":`, http.StatusBadRequest, "invalid json"},
		{"大文字と小文字だけ違うログインID", `{"loginId":"TARO_01","email":"other@example.com","password":"password"}`, http.StatusConflict, "login id taken"},
		{"登録済みのメールアドレス", `{"loginId":"hanako","email":"taro@example.com","password":"password"}`, http.StatusConflict, "email taken"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer(t)
			call(t, newClient(t), "POST", ts.url+"/api/signup", signupBody, nil)
			var body map[string]string
			if status := call(t, newClient(t), "POST", ts.url+"/api/signup", tt.body, &body); status != tt.status || body["error"] != tt.error {
				t.Fatalf("got (%d, %q), want (%d, %q)", status, body["error"], tt.status, tt.error)
			}
		})
	}
}

func TestLoginErrors(t *testing.T) {
	ts := newTestServer(t)
	call(t, newClient(t), "POST", ts.url+"/api/signup", signupBody, nil)
	// ログインID違いもパスワード違いも、同じ 401 にする。
	for _, body := range []string{
		`{"loginId":"taro_01","password":"wrong-password"}`,
		`{"loginId":"nobody","password":"password"}`,
	} {
		c := newClient(t)
		var res map[string]string
		if status := call(t, c, "POST", ts.url+"/api/login", body, &res); status != http.StatusUnauthorized || res["error"] != "invalid credentials" {
			t.Fatalf("%s: got (%d, %v)", body, status, res)
		}
		if status := call(t, c, "GET", ts.url+"/api/me", "", nil); status != http.StatusUnauthorized {
			t.Fatalf("%s: ログインできてしまう", body)
		}
	}
}

func TestPostRequiresJSON(t *testing.T) {
	ts := newTestServer(t)
	for _, path := range []string{"/api/login", "/api/questions/1/attempts"} {
		res, err := http.Post(ts.url+path, "application/x-www-form-urlencoded", strings.NewReader("loginId=taro"))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusUnsupportedMediaType {
			t.Fatalf("%s: status = %d, want %d", path, res.StatusCode, http.StatusUnsupportedMediaType)
		}
	}
}

func TestAttemptRecordsUser(t *testing.T) {
	ts := newTestServer(t)
	anonymous := newClient(t)
	call(t, anonymous, "POST", ts.url+"/api/questions/1/attempts", `{"answers":{}}`, nil)

	c := newClient(t)
	var u auth.User
	call(t, c, "POST", ts.url+"/api/signup", signupBody, &u)
	if status := call(t, c, "POST", ts.url+"/api/questions/1/attempts", `{"answers":{}}`, nil); status != http.StatusOK {
		t.Fatalf("attempt: status = %d", status)
	}

	saved := ts.questions.Attempts()
	if len(saved) != 2 || saved[0].UserID != nil || saved[1].UserID == nil || *saved[1].UserID != u.ID {
		t.Fatalf("saved attempts = %+v, want [未ログイン, user %d]", saved, u.ID)
	}
}

func TestUnknownSessionIsAnonymous(t *testing.T) {
	users := auth.NewMemoryUserStore()
	questions := quiz.NewMemoryStore(quiz.SampleQuestions()...)
	before := httptest.NewServer(New(Config{Questions: questions, Users: users, Sessions: auth.NewMemorySessionStore()}))
	t.Cleanup(before.Close)
	c := newClient(t)
	call(t, c, "POST", before.URL+"/api/signup", signupBody, nil)

	// セッションの置き場所が空になった（期限切れ、再起動）サーバーに、同じ Cookie を送る。
	after := httptest.NewServer(New(Config{Questions: questions, Users: users, Sessions: auth.NewMemorySessionStore()}))
	t.Cleanup(after.Close)
	c.Jar.SetCookies(mustParse(t, after.URL), c.Jar.Cookies(mustParse(t, before.URL)))

	if status := call(t, c, "GET", after.URL+"/api/me", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("/me: status = %d", status)
	}
	if status := call(t, c, "GET", after.URL+"/api/questions", "", nil); status != http.StatusOK {
		t.Fatalf("問題一覧は見られる: status = %d", status)
	}
}

func mustParse(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
