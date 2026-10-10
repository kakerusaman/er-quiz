package quiz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewHandler(NewMemoryStore(SampleQuestions()...)))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetQuestionHidesAnswers(t *testing.T) {
	srv := newTestServer(t)
	res, err := http.Get(srv.URL + "/api/questions/shop")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	for _, b := range raw["blanks"].([]any) {
		if _, ok := b.(map[string]any)["answer"]; ok {
			t.Fatalf("answer leaked: %v", b)
		}
	}
}

func TestGetQuestionNotFound(t *testing.T) {
	srv := newTestServer(t)
	res, err := http.Get(srv.URL + "/api/questions/nope")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestPostAnswers(t *testing.T) {
	srv := newTestServer(t)
	body := `{"answers":{"b1":"email","b2":"users","b3":"user_id"}}`
	res, err := http.Post(srv.URL+"/api/questions/shop/answers", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got Result
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Score != 2 || got.Total != 3 || got.Correct["b2"] {
		t.Fatalf("unexpected result: %+v", got)
	}
}
