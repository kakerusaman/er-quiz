package quiz

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kakerusaman/er-quiz/backend/internal/auth"
)

// Register は api グループに問題の一覧・取得・解答のルートを登録する。
// 解答はログインなしでもできる。ログイン中なら auth.Handler.LoadUser が読んだユーザーを記録する。
func Register(api *gin.RouterGroup, s Store) {
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.GET("/questions", func(c *gin.Context) {
		list, err := s.List(c.Request.Context())
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, list)
	})
	api.GET("/questions/:id", func(c *gin.Context) {
		q, ok := getQuestion(c, s)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, q.Public())
	})
	api.POST("/questions/:id/attempts", func(c *gin.Context) {
		q, ok := getQuestion(c, s)
		if !ok {
			return
		}
		var body struct {
			Answers map[int64]int64 `json:"answers"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		if err := q.Validate(body.Answers); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := q.Grade(body.Answers)
		a := Attempt{QuestionID: q.ID, Answers: body.Answers, Result: result}
		if u, ok := auth.CurrentUser(c); ok {
			a.UserID = &u.ID
		}
		id, err := s.SaveAttempt(c.Request.Context(), a)
		if err != nil {
			writeError(c, err)
			return
		}
		result.AttemptID = id
		c.JSON(http.StatusOK, result)
	})
}

// getQuestion はパスの :id の問題を取得する。失敗したらエラーを書いて false を返す。
func getQuestion(c *gin.Context, s Store) (QuestionData, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeError(c, ErrNotFound)
		return QuestionData{}, false
	}
	q, err := s.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return QuestionData{}, false
	}
	return q, true
}

func writeError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
		return
	}
	log.Printf("internal error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}
