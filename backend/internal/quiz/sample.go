package quiz

// SampleQuestions は動作確認用の問題。
func SampleQuestions() []Question {
	return []Question{{
		ID:    "shop",
		Title: "ECサイトの注文",
		Tables: []Table{
			{ID: "users", Name: "users", X: 0, Y: 0, Columns: []Column{
				{ID: "users.id", Name: "id", Type: "bigint", PK: true},
				{ID: "users.email", Type: "varchar", BlankID: "b1"},
			}},
			{ID: "orders", BlankID: "b2", X: 320, Y: 0, Columns: []Column{
				{ID: "orders.id", Name: "id", Type: "bigint", PK: true},
				{ID: "orders.user_id", Type: "bigint", FK: true, BlankID: "b3"},
				{ID: "orders.ordered_at", Name: "ordered_at", Type: "timestamp"},
			}},
		},
		Relations: []Relation{{ID: "r1", From: "users", To: "orders", Cardinality: "1:N"}},
		Blanks: []Blank{
			{ID: "b1", Choices: []string{"email", "order_id", "price"}, Answer: "email"},
			{ID: "b2", Choices: []string{"orders", "users", "items"}, Answer: "orders"},
			{ID: "b3", Choices: []string{"item_id", "user_id", "email"}, Answer: "user_id"},
		},
	}}
}
