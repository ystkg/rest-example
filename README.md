# REST APIサンプル

# 概要

- Go言語によるREST APIのサンプルコード
- 題材は商品の価格推移を記録するWebアプリケーション
- 実用性ではなくフレームワークの特徴を把握するのが目的
- 認証は `JWT`
- テーブルはアプリケーションの起動時に `GORM` のAuto Migrationで生成

<table>
    <tr><td> Webフレームワーク </td><td> Echo </td></tr>
    <tr><td> ORMライブラリ </td><td> GORM </td></tr>
    <tr><td> テストフレームワーク </td><td> Testify </td></tr>
    <tr><td> データベース </td><td> PostgreSQL </td></tr>
</table>

# API

## ユーザ

| 操作 | METHOD | ENDPOINT | STATUS CODE | REQUEST BODY | RESPONSE BODY |
| ---- | ---- | ---- | :----: | ---- | ---- |
| 登録         | POST | /user             | 201 | application/x-www-form-urlencoded | application/json |
| トークン発行 | POST | /user/:name/token | 201 | application/x-www-form-urlencoded | application/json |

## 価格

| 操作 | METHOD | ENDPOINT | STATUS CODE | REQUEST BODY | RESPONSE BODY |
| ---- | ---- | ---- | :----: | ---- | ---- |
| 登録 | POST   | /v1/price     | 201 | application/json | application/json |
| 一覧 | GET    | /v1/price     | 200 | -                | application/json |
| 取得 | GET    | /v1/price/:id | 200 | -                | application/json |
| 更新 | PUT    | /v1/price/:id | 200 | application/json | application/json |
| 削除 | DELETE | /v1/price/:id | 204 | -                | -                |

# エンティティ

```mermaid
erDiagram
    users ||--o{ prices : "登録する"
    users {
        uint id PK
        datetime created_at
        datetime updated_at
        datetime deleted_at
        string name UK
        string password
    }
    prices {
        uint id PK
        datetime created_at
        datetime updated_at
        datetime deleted_at
        uint user_id FK
        datetime date_time
        string store
        string product
        uint price
        bool in_stock
    }
```

# 使い方

## 起動方法

### 依存モジュールの準備

```Shell
go mod tidy
```

### PostgreSQLの起動

```Shell
docker compose up -d
```

- 上記はDocker Composeのプラグイン版。スタンドアローン版の場合は `docker compose` を `docker-compose` に置換

### 環境変数に接続先のデータベースを設定

```Shell
export DBURL=postgres://postgres:develop@localhost:5432/?sslmode=disable
```

### アプリケーションの起動

```Shell
go run .
```

## 実行例

### ユーザを登録

```Shell
curl -X POST -d 'name=user1' -d 'password=pw123' http://localhost:1323/user
```
```JSON
{
  "ID": 1,
  "name": "user1"
}
```

### トークンの発行

```Shell
curl -X POST -d 'password=pw123' http://localhost:1323/user/user1/token
```
```JSON
{
  "Token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODUxNzc2NTQsIlVzZXJJZCI6MX0.ypaxze8P9rIylG03y01E98n3YdED77P4x5xZo--TeIw"
}
```

- 有効期限は2時間

### シェル変数にトークンを設定

```Shell
TOKEN=`curl -s -X POST -d 'password=pw123' http://localhost:1323/user/user1/token | jq -r '.Token'`
```

### 価格を登録

```Shell
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"DateTime":"2023-05-15 12:10:30", "Store":"pcshop", "Product":"ssd1T", "Price":17800, "InStock":true}' http://localhost:1323/v1/price
```
```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 10:10:10",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17800,
  "InStock": true
}
```

### 価格の一覧

```Shell
curl -X GET -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/price
```
```JSON
[
  {
    "ID": 1,
    "DateTime": "2023-05-15 01:10:10",
    "Store": "pcshop",
    "Product": "ssd1T",
    "Price": 17800,
    "InStock": true
  }
]
```

### 価格の取得

```Shell
curl -X GET -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/price/1
```
```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 01:10:10",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17800,
  "InStock": true
}
```

### 価格の更新

```Shell
curl -X PUT -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"DateTime":"2023-05-15 12:10:30", "Store":"pcshop", "Product":"ssd1T", "Price":17500, "InStock":true}' http://localhost:1323/v1/price/1
```
```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 10:10:10",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17500,
  "InStock": true
}
```

### 価格の削除

```Shell
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/price/1
```

# 開発関連

## テスト

- テスト用に `PostgreSQL` を `15432` ポートで起動させた状態でテスト実行
- テスト毎に個別のデータベースを作成し、テストが成功したら削除。失敗時は確認用に最後に実行した状態を保持

### PostgreSQLの起動

```Shell
docker compose up -d
```

### テストの実行

```Shell
go test ./handler
```

## Docker

### イメージの作成

```Shell
docker build -t rest-example .
```

### PostgreSQLの起動

```Shell
docker compose up -d
```

### アプリケーションの起動

```Shell
docker run --rm --network=rest-develop -p 1323:1323 -e DBURL=postgres://postgres:develop@postgres-develop:5432/?sslmode=disable rest-example
```

## 環境変数

| 環境変数名 | 必須 | 説明 |
| ---- | :----: | ---- |
| DBURL | 〇 | PostgreSQLの接続文字列<br>例）<br> `postgres://postgres:develop@localhost:5432/?sslmode=disable` <br> _※テスト用は固定で変更不可_ |
| JWTKEY |  | 固定したい場合などに任意の文字列を指定。<br>省略した場合、アプリケーションの起動時にランダム生成し、<br>停止すると発行したトークンは有効期限前に **無効** になる |
| ECHOADDRESS |  | 省略時は `:1323` |
