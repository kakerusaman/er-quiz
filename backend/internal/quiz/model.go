// Package quiz はER図の穴埋め問題のモデル・保存・採点を扱う。
package quiz

// Column はテーブルのカラム。BlankID が空でなければ、その名前は空欄として出題される。
type Column struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type"`
	PK      bool   `json:"pk"`
	FK      bool   `json:"fk"`
	BlankID string `json:"blankId,omitempty"`
}

// Table はER図上のエンティティ。X, Y は図上の初期位置。
type Table struct {
	ID      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	BlankID string   `json:"blankId,omitempty"`
	X       float64  `json:"x"`
	Y       float64  `json:"y"`
	Columns []Column `json:"columns"`
}

// Relation はテーブル間のリレーション。Cardinality は "1:1" / "1:N" / "N:M"。
type Relation struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Cardinality string `json:"cardinality"`
}

// Blank は空欄1つ分。Answer はクライアントに返さない。
type Blank struct {
	ID      string   `json:"id"`
	Choices []string `json:"choices"`
	Answer  string   `json:"-"`
}

// Question は1問分のER図。
type Question struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Tables    []Table    `json:"tables"`
	Relations []Relation `json:"relations"`
	Blanks    []Blank    `json:"blanks"`
}

// Summary は一覧表示用の問題情報。
type Summary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Result は採点結果。
type Result struct {
	Correct map[string]bool `json:"correct"`
	Score   int             `json:"score"`
	Total   int             `json:"total"`
}

// Grade は空欄IDごとの回答を採点する。未回答は不正解として扱う。
func (q Question) Grade(answers map[string]string) Result {
	r := Result{Correct: make(map[string]bool, len(q.Blanks)), Total: len(q.Blanks)}
	for _, b := range q.Blanks {
		ok := answers[b.ID] == b.Answer
		r.Correct[b.ID] = ok
		if ok {
			r.Score++
		}
	}
	return r
}
