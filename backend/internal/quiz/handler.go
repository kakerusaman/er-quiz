package quiz

import (
	"encoding/json"
	"errors"
	"net/http"
)

// NewHandler は /api 以下のルートを持つ http.Handler を返す。
func NewHandler(s Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/questions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List())
	})
	mux.HandleFunc("GET /api/questions/{id}", func(w http.ResponseWriter, r *http.Request) {
		q, err := s.Get(r.PathValue("id"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, q)
	})
	mux.HandleFunc("POST /api/questions/{id}/answers", func(w http.ResponseWriter, r *http.Request) {
		q, err := s.Get(r.PathValue("id"))
		if err != nil {
			writeError(w, err)
			return
		}
		var body struct {
			Answers map[string]string `json:"answers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		writeJSON(w, http.StatusOK, q.Grade(body.Answers))
	})
	return mux
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
