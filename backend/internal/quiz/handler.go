package quiz

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// NewHandler は /api 以下のルートを持つ http.Handler を返す。
func NewHandler(s Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/questions", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.List(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	})
	mux.HandleFunc("GET /api/questions/{id}", func(w http.ResponseWriter, r *http.Request) {
		q, ok := getQuestion(w, r, s)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, q.Public())
	})
	mux.HandleFunc("POST /api/questions/{id}/attempts", func(w http.ResponseWriter, r *http.Request) {
		q, ok := getQuestion(w, r, s)
		if !ok {
			return
		}
		var body struct {
			Answers map[int64]int64 `json:"answers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		if err := q.Validate(body.Answers); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		result := q.Grade(body.Answers)
		id, err := s.SaveAttempt(r.Context(), Attempt{QuestionID: q.ID, Answers: body.Answers, Result: result})
		if err != nil {
			writeError(w, err)
			return
		}
		result.AttemptID = id
		writeJSON(w, http.StatusOK, result)
	})
	return mux
}

// getQuestion はパスの {id} の問題を取得する。失敗したらエラーを書いて false を返す。
func getQuestion(w http.ResponseWriter, r *http.Request, s Store) (QuestionData, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrNotFound)
		return QuestionData{}, false
	}
	q, err := s.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return QuestionData{}, false
	}
	return q, true
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": ErrNotFound.Error()})
		return
	}
	log.Printf("internal error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
