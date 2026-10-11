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

func TestPostAnswers(t *testing.T) {
	srv := newTestServer(t)
	// ア=注文(正解), イ=会員(不正解), ウ=未回答
	body := `{"answers":{"1":1,"2":2}}`
	res, err := http.Post(srv.URL+"/api/questions/1/answers", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got Result
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Score != 1 || got.Total != 3 || !got.Correct[1] || got.Correct[2] || got.Correct[3] {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestPublicDoesNotModifyOriginal(t *testing.T) {
	d := SampleQuestions()[0]
	_ = d.Public()
	if d.Entities[1].Name == nil || d.Entities[0].Attributes[2].Name == nil {
		t.Fatal("Public() cleared names in the original data")
	}
}
