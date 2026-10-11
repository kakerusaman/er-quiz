package quiz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PostgresStore は PostgreSQL に保存された問題を読む Store。
// テーブル定義は backend/db/schema.sql（docs/design/data.md）を参照。
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore は接続済みの *sql.DB から PostgresStore を作る。
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) List(ctx context.Context) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT q.id, q.title, count(b.id)
		FROM questions q
		LEFT JOIN blanks b ON b.question_id = q.id
		GROUP BY q.id
		ORDER BY q.created_at DESC, q.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	defer rows.Close()
	out := []Summary{}
	for rows.Next() {
		var sm Summary
		if err := rows.Scan(&sm.ID, &sm.Title, &sm.BlankCount); err != nil {
			return nil, fmt.Errorf("list questions: %w", err)
		}
		out = append(out, sm)
	}
	return out, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id int64) (QuestionData, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return QuestionData{}, fmt.Errorf("get question %d: %w", id, err)
	}
	defer tx.Rollback() //nolint:errcheck // 読み取り専用なので結果は使わない

	d := QuestionData{Answers: map[int64]int64{}}
	err = tx.QueryRowContext(ctx,
		`SELECT id, title, scenario, explanation FROM questions WHERE id = $1`, id,
	).Scan(&d.ID, &d.Title, &d.Scenario, &d.Explanation)
	if errors.Is(err, sql.ErrNoRows) {
		return QuestionData{}, ErrNotFound
	}
	if err != nil {
		return QuestionData{}, fmt.Errorf("get question %d: %w", id, err)
	}

	// 空欄: エンティティIDまたは属性ID → 空欄
	entityBlanks := map[int64]*BlankRef{}
	attrBlanks := map[int64]*BlankRef{}
	if err := eachRow(ctx, tx, `
		SELECT id, symbol, entity_id, attribute_id, answer_choice_id
		FROM blanks WHERE question_id = $1 ORDER BY id`, []any{id},
		func(rows *sql.Rows) error {
			var b BlankRef
			var entityID, attrID sql.NullInt64
			var answer int64
			if err := rows.Scan(&b.ID, &b.Symbol, &entityID, &attrID, &answer); err != nil {
				return err
			}
			if entityID.Valid {
				entityBlanks[entityID.Int64] = &b
			} else {
				attrBlanks[attrID.Int64] = &b
			}
			d.Answers[b.ID] = answer
			return nil
		}); err != nil {
		return QuestionData{}, fmt.Errorf("get question %d blanks: %w", id, err)
	}

	entityIndex := map[int64]int{}
	d.Entities = []Entity{}
	if err := eachRow(ctx, tx, `
		SELECT id, name, x, y FROM entities WHERE question_id = $1 ORDER BY id`, []any{id},
		func(rows *sql.Rows) error {
			var e Entity
			var name string
			if err := rows.Scan(&e.ID, &name, &e.X, &e.Y); err != nil {
				return err
			}
			e.Name = &name
			e.Blank = entityBlanks[e.ID]
			e.Attributes = []Attribute{}
			entityIndex[e.ID] = len(d.Entities)
			d.Entities = append(d.Entities, e)
			return nil
		}); err != nil {
		return QuestionData{}, fmt.Errorf("get question %d entities: %w", id, err)
	}

	if err := eachRow(ctx, tx, `
		SELECT a.id, a.entity_id, a.name, a.is_pk, a.is_fk
		FROM attributes a JOIN entities e ON e.id = a.entity_id
		WHERE e.question_id = $1 ORDER BY a.entity_id, a.position`, []any{id},
		func(rows *sql.Rows) error {
			var a Attribute
			var entityID int64
			var name string
			if err := rows.Scan(&a.ID, &entityID, &name, &a.IsPK, &a.IsFK); err != nil {
				return err
			}
			a.Name = &name
			a.Blank = attrBlanks[a.ID]
			e := &d.Entities[entityIndex[entityID]]
			e.Attributes = append(e.Attributes, a)
			return nil
		}); err != nil {
		return QuestionData{}, fmt.Errorf("get question %d attributes: %w", id, err)
	}

	d.Relationships = []Relationship{}
	if err := eachRow(ctx, tx, `
		SELECT id, from_entity_id, to_entity_id, from_cardinality, to_cardinality
		FROM relationships WHERE question_id = $1 ORDER BY id`, []any{id},
		func(rows *sql.Rows) error {
			var r Relationship
			if err := rows.Scan(&r.ID, &r.FromEntityID, &r.ToEntityID, &r.FromCardinality, &r.ToCardinality); err != nil {
				return err
			}
			d.Relationships = append(d.Relationships, r)
			return nil
		}); err != nil {
		return QuestionData{}, fmt.Errorf("get question %d relationships: %w", id, err)
	}

	d.Choices = []Choice{}
	if err := eachRow(ctx, tx, `
		SELECT id, label FROM choices WHERE question_id = $1 ORDER BY position`, []any{id},
		func(rows *sql.Rows) error {
			var c Choice
			if err := rows.Scan(&c.ID, &c.Label); err != nil {
				return err
			}
			d.Choices = append(d.Choices, c)
			return nil
		}); err != nil {
		return QuestionData{}, fmt.Errorf("get question %d choices: %w", id, err)
	}

	return d, nil
}

// eachRow は query の結果の各行に fn を呼ぶ。
func eachRow(ctx context.Context, tx *sql.Tx, query string, args []any, fn func(*sql.Rows) error) error {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
