-- サンプル問題。schema.sql の後に流す。
-- 1問目は backend/internal/quiz/sample.go と同じ内容。

BEGIN;

-- 1問目: ECサイトの注文 ------------------------------------------------------

INSERT INTO questions (id, title, scenario, explanation, created_at) OVERRIDING SYSTEM VALUE VALUES
(1, 'ECサイトの注文',
 'ネットショップを運営するA社では、会員登録をした顧客が商品を注文する。'
 '会員は会員番号で識別し、氏名とメールアドレスを登録する。会員は、まだ注文していなくてもよい。'
 '注文は注文番号で識別し、注文日と注文した会員を記録する。1回の注文には、1件以上の注文明細が含まれる。'
 '注文明細には、注文した商品と数量を記録する。商品は商品コードで識別し、商品名と単価を管理する。'
 '1つの商品は、複数の注文明細に登場できる。',
 '「会員」と「注文」は1対多の関係で、会員は注文を0件以上持てます。'
 '注文と注文明細も1対多で、1回の注文には1件以上の明細が含まれます。'
 '商品は複数の注文明細に登場できるため、注文明細は商品コードを持って商品と結びつきます。'
 '会員が登録する項目は会員番号・氏名・メールアドレスです。',
 '2026-01-01 00:00:00+09');

INSERT INTO entities (id, question_id, name, x, y) OVERRIDING SYSTEM VALUE VALUES
(1, 1, '会員',     0,   0),
(2, 1, '注文',     340, 0),
(3, 1, '注文明細', 680, 0),
(4, 1, '商品',     680, 240);

INSERT INTO attributes (id, entity_id, name, is_pk, is_fk, position) OVERRIDING SYSTEM VALUE VALUES
(1,  1, '会員番号',       true,  false, 1),
(2,  1, '氏名',           false, false, 2),
(3,  1, 'メールアドレス', false, false, 3),
(4,  2, '注文番号',       true,  false, 1),
(5,  2, '会員番号',       false, true,  2),
(6,  2, '注文日',         false, false, 3),
(7,  3, '注文番号',       true,  true,  1),
(8,  3, '商品コード',     true,  true,  2),
(9,  3, '数量',           false, false, 3),
(10, 4, '商品コード',     true,  false, 1),
(11, 4, '商品名',         false, false, 2),
(12, 4, '単価',           false, false, 3);

INSERT INTO relationships (id, question_id, from_entity_id, to_entity_id, from_cardinality, to_cardinality) OVERRIDING SYSTEM VALUE VALUES
(1, 1, 1, 2, '1', '0..N'),
(2, 1, 2, 3, '1', '1..N'),
(3, 1, 4, 3, '1', '0..N');

INSERT INTO choices (id, question_id, label, position) OVERRIDING SYSTEM VALUE VALUES
(1, 1, '注文',           1),
(2, 1, '会員',           2),
(3, 1, '商品',           3),
(4, 1, 'メールアドレス', 4),
(5, 1, '商品コード',     5),
(6, 1, '配送先',         6),
(7, 1, '在庫数',         7);

INSERT INTO blanks (id, question_id, symbol, entity_id, attribute_id, answer_choice_id) OVERRIDING SYSTEM VALUE VALUES
(1, 1, 'ア', 2,    NULL, 1),
(2, 1, 'イ', NULL, 3,    4),
(3, 1, 'ウ', NULL, 8,    5);

-- 2問目: 図書館の貸出 --------------------------------------------------------

INSERT INTO questions (id, title, scenario, explanation, created_at) OVERRIDING SYSTEM VALUE VALUES
(2, '図書館の貸出',
 'B市の図書館では、利用者が会員登録をして資料を借りる。'
 '会員は会員番号で識別し、氏名と電話番号を登録する。資料は資料番号で識別し、書名と著者名を管理する。'
 '1回の貸出では、1人の会員が1冊の資料を借りる。貸出は貸出番号で識別し、貸出日と返却予定日を記録する。'
 '会員は、まだ一度も借りていなくてもよい。1冊の資料は、時期を変えて何度でも貸し出される。',
 '「会員」と「貸出」、「資料」と「貸出」は、どちらも1対多の関係です。'
 '貸出は、誰が（会員番号）どの資料を（資料番号）借りたかを外部キーとして持ちます。'
 '貸出ごとに変わる貸出日と返却予定日は、会員や資料ではなく貸出に持たせます。',
 '2026-01-02 00:00:00+09');

INSERT INTO entities (id, question_id, name, x, y) OVERRIDING SYSTEM VALUE VALUES
(5, 2, '会員', 0,   0),
(6, 2, '貸出', 340, 0),
(7, 2, '資料', 680, 0);

INSERT INTO attributes (id, entity_id, name, is_pk, is_fk, position) OVERRIDING SYSTEM VALUE VALUES
(13, 5, '会員番号',   true,  false, 1),
(14, 5, '氏名',       false, false, 2),
(15, 5, '電話番号',   false, false, 3),
(16, 6, '貸出番号',   true,  false, 1),
(17, 6, '会員番号',   false, true,  2),
(18, 6, '資料番号',   false, true,  3),
(19, 6, '貸出日',     false, false, 4),
(20, 6, '返却予定日', false, false, 5),
(21, 7, '資料番号',   true,  false, 1),
(22, 7, '書名',       false, false, 2),
(23, 7, '著者名',     false, false, 3);

INSERT INTO relationships (id, question_id, from_entity_id, to_entity_id, from_cardinality, to_cardinality) OVERRIDING SYSTEM VALUE VALUES
(4, 2, 5, 6, '1', '0..N'),
(5, 2, 7, 6, '1', '0..N');

INSERT INTO choices (id, question_id, label, position) OVERRIDING SYSTEM VALUE VALUES
(8,  2, '貸出',       1),
(9,  2, '会員',       2),
(10, 2, '予約',       3),
(11, 2, '電話番号',   4),
(12, 2, '資料番号',   5),
(13, 2, '書名',       6),
(14, 2, '返却予定日', 7),
(15, 2, '住所',       8),
(16, 2, '出版社',     9);

INSERT INTO blanks (id, question_id, symbol, entity_id, attribute_id, answer_choice_id) OVERRIDING SYSTEM VALUE VALUES
(4, 2, 'ア', 6,    NULL, 8),
(5, 2, 'イ', NULL, 15,   11),
(6, 2, 'ウ', NULL, 18,   12),
(7, 2, 'エ', NULL, 22,   13),
(8, 2, 'オ', NULL, 20,   14);

-- 明示した ID の続きから採番されるようにする
SELECT setval(pg_get_serial_sequence('questions', 'id'),     (SELECT max(id) FROM questions));
SELECT setval(pg_get_serial_sequence('entities', 'id'),      (SELECT max(id) FROM entities));
SELECT setval(pg_get_serial_sequence('attributes', 'id'),    (SELECT max(id) FROM attributes));
SELECT setval(pg_get_serial_sequence('relationships', 'id'), (SELECT max(id) FROM relationships));
SELECT setval(pg_get_serial_sequence('choices', 'id'),       (SELECT max(id) FROM choices));
SELECT setval(pg_get_serial_sequence('blanks', 'id'),        (SELECT max(id) FROM blanks));

COMMIT;
