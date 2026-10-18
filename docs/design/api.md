# API設計

[仕様](../spec.md)・[データ設計](data.md)・[画面設計](screen.md) をもとにした API 設計です。MVP とログインの範囲を対象にしています。

## 方針

- 正解（どの選択肢が正解か）と解説は、答え合わせのレスポンスでだけ返す。問題取得では返さない。
- 空欄になっているエンティティ名・属性名は、問題取得では `name` を返さない（null）。
- 回答は選択肢の文字列ではなく、選択肢の ID で送る（同じ文字列の表記ゆれを気にしなくてよい）。
- ID は数値。JSON のキーは camelCase。
- エラーは `{"error": "..."}` の形で返す。`error` は英語の決まった文字列で、画面ではこれを日本語の文に直して出す。
- POST は `Content-Type: application/json` だけを受け付け、それ以外は 415 を返す（他サイトのフォームから送らせる CSRF を防ぐ）。

## エンドポイント

| メソッド | パス | 内容 |
| --- | --- | --- |
| GET | `/api/health` | 死活確認（変更なし） |
| GET | `/api/questions` | 問題一覧 |
| GET | `/api/questions/{id}` | 問題本体（正解・解説なし） |
| POST | `/api/questions/{id}/attempts` | 回答を採点し、記録して、結果を返す |
| POST | `/api/signup` | 新規登録して、そのままログインする |
| POST | `/api/login` | ログインする |
| POST | `/api/logout` | ログアウトする |
| GET | `/api/me` | ログイン中のユーザー |

## GET /api/questions

```json
[
  { "id": 1, "title": "ECサイトの注文", "blankCount": 3 }
]
```

新しく追加した問題が上（questions.created_at の降順）。

## GET /api/questions/{id}

```json
{
  "id": 1,
  "title": "ECサイトの注文",
  "scenario": "〇〇社では、会員が商品を注文する。…",
  "entities": [
    {
      "id": 10, "name": "会員", "x": 0, "y": 0, "blank": null,
      "attributes": [
        { "id": 100, "name": "会員番号", "isPk": true, "isFk": false, "blank": null },
        { "id": 101, "name": null, "isPk": false, "isFk": false, "blank": { "id": 1000, "symbol": "イ" } }
      ]
    },
    {
      "id": 11, "name": null, "x": 300, "y": 0, "blank": { "id": 1001, "symbol": "ア" },
      "attributes": []
    }
  ],
  "relationships": [
    { "id": 20, "fromEntityId": 10, "toEntityId": 11, "fromCardinality": "1", "toCardinality": "0..N" }
  ],
  "choices": [
    { "id": 500, "label": "注文" },
    { "id": 501, "label": "メールアドレス" }
  ]
}
```

- 存在しない id は 404。

## POST /api/questions/{id}/attempts

リクエスト（空欄 ID → 置いた選択肢 ID。未回答の空欄は省略してよい）

```json
{ "answers": { "1000": 501, "1001": 500 } }
```

レスポンス

```json
{
  "attemptId": 42,
  "score": 2,
  "total": 2,
  "results": [
    { "blankId": 1000, "correct": true,  "answerChoiceId": 501 },
    { "blankId": 1001, "correct": true,  "answerChoiceId": 500 }
  ],
  "explanation": "「会員」と「注文」は1対多の関係で…"
}
```

- `answerChoiceId` は正解の選択肢。画面では不正解の空欄の下に「正解: ○○」としてこの選択肢の文字列を出す。
- 解答結果は attempts / attempt_answers に保存する。ログインなしでも送れる。ログイン中なら attempts.user_id にその人を記録する。未回答の空欄も choice_id を NULL にして記録する。「もう一度解く」は新しい解答として送る。
- 存在しない問題は 404。この問題にない空欄 ID・選択肢 ID が含まれていたら 400。同じ選択肢 ID を複数の空欄に使っていても 400。

## ログイン

ログイン状態は Cookie `session`（HttpOnly、SameSite=Lax、30日。`SECURE_COOKIE=true` のとき Secure）で持つ。中身はランダムなトークンで、サーバーは Redis でユーザーを引く（[データ設計](data.md) の「セッション（Redis）」）。

ユーザーの JSON（パスワードのハッシュは返さない）

```json
{ "id": 1, "loginId": "taro", "email": "taro@example.com" }
```

### POST /api/signup

```json
{ "loginId": "Taro", "email": "taro@example.com", "password": "password" }
```

- 成功すると 201 とユーザーを返し、Cookie を付ける。ログインIDとメールアドレスは小文字にそろえて保存する。
- 形式の誤りは 400（`invalid login id` / `invalid email` / `invalid password`）。
- ほかの人が使っていれば 409（`login id taken` / `email taken`）。

### POST /api/login

```json
{ "loginId": "taro", "password": "password" }
```

- 成功すると 200 とユーザーを返し、Cookie を付ける。それまでのセッションは消す。
- ログインIDかパスワードが違えば 401（`invalid credentials`）。どちらが違うかは返さない。

### POST /api/logout

- セッションを消し、Cookie も消して 204 を返す。ログインしていなくても 204。

### GET /api/me

- ログイン中なら 200 とユーザー、未ログインなら 401（`login required`）。
- ログインが必要な API には、これと同じミドルウェア（`auth.RequireLogin`）を付ける。
