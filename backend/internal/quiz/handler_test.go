package quiz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewHandler(NewMemoryStore(SampleQuestions()...)))
	t.Cleanup(srv.Close)
	return srv
}

func getJSON(t *testing.T, url string, v any) int {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if v != nil && res.StatusCode == http.StatusOK {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			t.Fatal(err)
		}
	}
	return res.StatusCode
}

func TestListQuestions(t *testing.T) {
	second := SampleQuestions()[0]
	second.ID, second.Title = 2, "新しい問題"
	srv := httptest.NewServer(NewHandler(NewMemoryStore(SampleQuestions()[0], second)))
	t.Cleanup(srv.Close)

	var got []Summary
	if status := getJSON(t, srv.URL+"/api/questions", &got); status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	want := []Summary{{ID: 2, Title: "新しい問題", BlankCount: 3}, {ID: 1, Title: "ECサイトの注文", BlankCount: 3}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v, want %+v（新しい問題が上）", got, want)
	}
}

func TestGetQuestionHidesAnswers(t *testing.T) {
	srv := newTestServer(t)
	res, err := http.Get(srv.URL + "/api/questions/1")
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
	for _, key := range []string{"answers", "Answers", "explanation", "Explanation"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("%q leaked", key)
		}
	}

	// 空欄の名前は null、それ以外は名前を返す。
	blankNames, shownNames := 0, 0
	for _, e := range raw["entities"].([]any) {
		e := e.(map[string]any)
		nodes := append([]any{e}, e["attributes"].([]any)...)
		for _, n := range nodes {
			n := n.(map[string]any)
			if n["blank"] != nil {
				blankNames++
				if n["name"] != nil {
					t.Fatalf("blank name leaked: %v", n)
				}
			} else {
				shownNames++
				if n["name"] == nil {
					t.Fatalf("name missing: %v", n)
				}
			}
		}
	}
	if blankNames != 3 || shownNames == 0 {
		t.Fatalf("blanks = %d, shown = %d", blankNames, shownNames)
	}
}

func TestGetQuestionNotFound(t *testing.T) {
	srv := newTestServer(t)
	for _, id := range []string{"999", "shop"} {
		if status := getJSON(t, srv.URL+"/api/questions/"+id, nil); status != http.StatusNotFound {
			t.Fatalf("id %s: status = %d", id, status)
		}
	}
}

func postAttempt(t *testing.T, url, body string) *http.Response {
	t.Helper()
	res, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func TestPostAttempt(t *testing.T) {
	store := NewMemoryStore(SampleQuestions()...)
	srv := httptest.NewServer(NewHandler(store))
	t.Cleanup(srv.Close)

	// ア=注文(正解), イ=会員(不正解), ウ=未回答
	res := postAttempt(t, srv.URL+"/api/questions/1/attempts", `{"answers":{"1":1,"2":2}}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	var got Result
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	want := Result{
		AttemptID: 1,
		Score:     1,
		Total:     3,
		Results: []BlankResult{
			{BlankID: 1, Correct: true, AnswerChoiceID: 1},
			{BlankID: 2, Correct: false, AnswerChoiceID: 4},
			{BlankID: 3, Correct: false, AnswerChoiceID: 5},
		},
		Explanation: SampleQuestions()[0].Explanation,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}

	saved := store.Attempts()
	if len(saved) != 1 || saved[0].QuestionID != 1 || !reflect.DeepEqual(saved[0].Answers, map[int64]int64{1: 1, 2: 2}) {
		t.Fatalf("saved attempts = %+v", saved)
	}

	// もう一度解くと別の解答として記録される。
	res = postAttempt(t, srv.URL+"/api/questions/1/attempts", `{"answers":{}}`)
	var again Result
	if err := json.NewDecoder(res.Body).Decode(&again); err != nil {
		t.Fatal(err)
	}
	if again.AttemptID != 2 || again.Score != 0 || len(store.Attempts()) != 2 {
		t.Fatalf("second attempt = %+v", again)
	}
}

func TestPostAttemptErrors(t *testing.T) {
	tests := []struct {
		name, path, body string
		status           int
	}{
		{"存在しない問題", "/api/questions/999/attempts", `{"answers":{}}`, http.StatusNotFound},
		{"数字でない問題ID", "/api/questions/shop/attempts", `{"answers":{}}`, http.StatusNotFound},
		{"他の問題の空欄", "/api/questions/1/attempts", `{"answers":{"4":1}}`, http.StatusBadRequest},
		{"他の問題の選択肢", "/api/questions/1/attempts", `{"answers":{"1":99}}`, http.StatusBadRequest},
		{"同じ選択肢を2つの空欄に使う", "/api/questions/1/attempts", `{"answers":{"1":1,"2":1}}`, http.StatusBadRequest},
		{"壊れたJSON", "/api/questions/1/attempts", `{"answers":`, http.StatusBadRequest},
		{"数字でない空欄ID", "/api/questions/1/attempts", `{"answers":{"ア":1}}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore(SampleQuestions()...)
			srv := httptest.NewServer(NewHandler(store))
			t.Cleanup(srv.Close)
			res := postAttempt(t, srv.URL+tt.path, tt.body)
			if res.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.status)
			}
			if n := len(store.Attempts()); n != 0 {
				t.Fatalf("%d attempts saved on error", n)
			}
		})
	}
}

func TestPublicDoesNotModifyOriginal(t *testing.T) {
	d := SampleQuestions()[0]
	_ = d.Public()
	if d.Entities[1].Name == nil || d.Entities[0].Attributes[2].Name == nil {
		t.Fatal("Public() cleared names in the original data")
	}
}
