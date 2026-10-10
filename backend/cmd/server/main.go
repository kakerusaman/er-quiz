package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kakerusaman/er-quiz/backend/internal/quiz"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	h := quiz.NewHandler(quiz.NewMemoryStore(quiz.SampleQuestions()...))
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}
