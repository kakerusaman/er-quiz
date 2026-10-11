// Package quiz はER図の穴埋め問題のモデル・保存・採点を扱う。
package quiz

// BlankRef は空欄の識別子と記号（ア、イ…）。正解は含まない。
type BlankRef struct {
	ID     int64  `json:"id"`
	Symbol string `json:"symbol"`
}

// Attribute はエンティティの属性。空欄なら Blank が入り、API では Name を返さない。
type Attribute struct {
	ID    int64     `json:"id"`
	Name  *string   `json:"name"`
	IsPK  bool      `json:"isPk"`
	IsFK  bool      `json:"isFk"`
	Blank *BlankRef `json:"blank"`
}

// Entity はER図上のエンティティ。X, Y は図上の位置（固定）。
type Entity struct {
	ID         int64       `json:"id"`
	Name       *string     `json:"name"`
	X          int         `json:"x"`
	Y          int         `json:"y"`
	Blank      *BlankRef   `json:"blank"`
	Attributes []Attribute `json:"attributes"`
}

// 多重度。IE記法の線の端に対応する。
const (
	CardinalityOne        = "1"
	CardinalityZeroOrOne  = "0..1"
	CardinalityOneOrMany  = "1..N"
	CardinalityZeroOrMany = "0..N"
)

// Relationship はエンティティ間のリレーションシップ。両端の多重度を持つ。
type Relationship struct {
	ID              int64  `json:"id"`
	FromEntityID    int64  `json:"fromEntityId"`
	ToEntityID      int64  `json:"toEntityId"`
	FromCardinality string `json:"fromCardinality"`
	ToCardinality   string `json:"toCardinality"`
}

// Choice は解答群の選択肢。
type Choice struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

// Question は API で返す問題本体。正解と解説は含まない。
type Question struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Scenario      string         `json:"scenario"`
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Choices       []Choice       `json:"choices"`
}

// Summary は一覧表示用の問題情報。
type Summary struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	BlankCount int    `json:"blankCount"`
}

// QuestionData は保存されている問題の全体。空欄の名前、正解、解説を含むので、
// API で返すときは必ず Public() を通す。
type QuestionData struct {
	Question
	Explanation string
	// Answers は空欄ID → 正解の選択肢ID。
	Answers map[int64]int64
}

// Public は空欄の名前を消した、API で返してよい問題を返す。元の値は変更しない。
func (d QuestionData) Public() Question {
	q := d.Question
	q.Entities = make([]Entity, len(d.Entities))
	for i, e := range d.Entities {
		if e.Blank != nil {
			e.Name = nil
		}
		attrs := make([]Attribute, len(e.Attributes))
		for j, a := range e.Attributes {
			if a.Blank != nil {
				a.Name = nil
			}
			attrs[j] = a
		}
		e.Attributes = attrs
		q.Entities[i] = e
	}
	return q
}

// Summary は一覧表示用の情報を返す。
func (d QuestionData) Summary() Summary {
	return Summary{ID: d.ID, Title: d.Title, BlankCount: len(d.Answers)}
}

// Result は採点結果。Correct のキーは空欄ID。
type Result struct {
	Correct map[int64]bool `json:"correct"`
	Score   int            `json:"score"`
	Total   int            `json:"total"`
}

// Grade は空欄ID → 選択肢ID の回答を採点する。未回答は不正解として扱う。
func (d QuestionData) Grade(answers map[int64]int64) Result {
	r := Result{Correct: make(map[int64]bool, len(d.Answers)), Total: len(d.Answers)}
	for blankID, want := range d.Answers {
		got, ok := answers[blankID]
		correct := ok && got == want
		r.Correct[blankID] = correct
		if correct {
			r.Score++
		}
	}
	return r
}
