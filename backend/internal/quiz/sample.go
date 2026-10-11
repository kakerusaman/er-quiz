package quiz

// SampleQuestions はテストと、DATABASE_URL を指定せずに起動したときに使う問題。
// 内容は backend/db/seed.sql の1問目と同じ。
func SampleQuestions() []QuestionData {
	s := func(v string) *string { return &v }
	return []QuestionData{{
		Question: Question{
			ID:    1,
			Title: "ECサイトの注文",
			Scenario: "ネットショップを運営するA社では、会員登録をした顧客が商品を注文する。" +
				"会員は会員番号で識別し、氏名とメールアドレスを登録する。会員は、まだ注文していなくてもよい。" +
				"注文は注文番号で識別し、注文日と注文した会員を記録する。1回の注文には、1件以上の注文明細が含まれる。" +
				"注文明細には、注文した商品と数量を記録する。商品は商品コードで識別し、商品名と単価を管理する。" +
				"1つの商品は、複数の注文明細に登場できる。",
			Entities: []Entity{
				{ID: 1, Name: s("会員"), X: 0, Y: 0, Attributes: []Attribute{
					{ID: 1, Name: s("会員番号"), IsPK: true},
					{ID: 2, Name: s("氏名")},
					{ID: 3, Name: s("メールアドレス"), Blank: &BlankRef{ID: 2, Symbol: "イ"}},
				}},
				{ID: 2, Name: s("注文"), X: 340, Y: 0, Blank: &BlankRef{ID: 1, Symbol: "ア"}, Attributes: []Attribute{
					{ID: 4, Name: s("注文番号"), IsPK: true},
					{ID: 5, Name: s("会員番号"), IsFK: true},
					{ID: 6, Name: s("注文日")},
				}},
				{ID: 3, Name: s("注文明細"), X: 680, Y: 0, Attributes: []Attribute{
					{ID: 7, Name: s("注文番号"), IsPK: true, IsFK: true},
					{ID: 8, Name: s("商品コード"), IsPK: true, IsFK: true, Blank: &BlankRef{ID: 3, Symbol: "ウ"}},
					{ID: 9, Name: s("数量")},
				}},
				{ID: 4, Name: s("商品"), X: 680, Y: 240, Attributes: []Attribute{
					{ID: 10, Name: s("商品コード"), IsPK: true},
					{ID: 11, Name: s("商品名")},
					{ID: 12, Name: s("単価")},
				}},
			},
			Relationships: []Relationship{
				{ID: 1, FromEntityID: 1, ToEntityID: 2, FromCardinality: CardinalityOne, ToCardinality: CardinalityZeroOrMany},
				{ID: 2, FromEntityID: 2, ToEntityID: 3, FromCardinality: CardinalityOne, ToCardinality: CardinalityOneOrMany},
				{ID: 3, FromEntityID: 4, ToEntityID: 3, FromCardinality: CardinalityOne, ToCardinality: CardinalityZeroOrMany},
			},
			Choices: []Choice{
				{ID: 1, Label: "注文"},
				{ID: 2, Label: "会員"},
				{ID: 3, Label: "商品"},
				{ID: 4, Label: "メールアドレス"},
				{ID: 5, Label: "商品コード"},
				{ID: 6, Label: "配送先"},
				{ID: 7, Label: "在庫数"},
			},
		},
		Explanation: "「会員」と「注文」は1対多の関係で、会員は注文を0件以上持てます。" +
			"注文と注文明細も1対多で、1回の注文には1件以上の明細が含まれます。" +
			"商品は複数の注文明細に登場できるため、注文明細は商品コードを持って商品と結びつきます。" +
			"会員が登録する項目は会員番号・氏名・メールアドレスです。",
		Answers: map[int64]int64{1: 1, 2: 4, 3: 5},
	}}
}
