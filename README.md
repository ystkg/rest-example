# REST APIアプリケーション

## 概要

- GoによるREST APIの実装モデル
- データベースは環境変数の値でPostgreSQLとMySQLのどちらかを選択
- テーブルは起動時に `GORM` のAuto Migrationで生成
- 認証は `JWT`
- 題材は商品の価格推移を記録していくWebアプリケーション

<table>
<tr><td> Webフレームワーク </td><td> Echo </td></tr>
<tr><td> ORMライブラリ </td><td> GORM </td></tr>
<tr><td> データベース </td><td> PostgreSQL or MySQL </td></tr>
<tr><td> テストフレームワーク </td><td> Testify </td></tr>
<tr><td> テストモック </td><td> go-sqlmock </td></tr>
<tr><td> テストコンテナ </td><td> testcontainers-go <br> dockertest </td></tr>
</table>

## API

### ユーザ

| 操作 | METHOD | ENDPOINT | STATUS CODE | REQUEST BODY | RESPONSE BODY |
| ---- | ---- | ---- | :----: | ---- | ---- |
| 登録         | POST | /users             | 201 | application/x-www-form-urlencoded | application/json |
| トークン発行 | POST | /users/:name/token | 201 | application/x-www-form-urlencoded | application/json |

### 価格

| 操作 | METHOD | ENDPOINT | STATUS CODE | REQUEST BODY | RESPONSE BODY |
| ---- | ---- | ---- | :----: | ---- | ---- |
| 登録 | POST   | /v1/prices     | 201 | application/json | application/json |
| 一覧 | GET    | /v1/prices     | 200 | -                | application/json |
| 取得 | GET    | /v1/prices/:id | 200 | -                | application/json |
| 更新 | PUT    | /v1/prices/:id | 200 | application/json | application/json |
| 削除 | DELETE | /v1/prices/:id | 204 | -                | -                |

## エンティティ

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
    }
```

## 使い方

### 動作要件

Go（version 1.23以降）がインストールされていること

例

```ShellSession
$ go version
go version go1.23.5 linux/amd64
```

### 起動方法

#### 依存モジュールの準備

```Shell
go mod tidy
```

#### データベースのコンテナ起動

Docker Composeはプラグイン版

##### PostgreSQL

```Shell
docker compose up -d --wait postgres
```

##### MySQL

```Shell
docker compose up -d --wait mysql
```

#### 環境変数に接続先のデータベースを設定

##### PostgreSQL使用時

```Shell
export DBURL=postgres://devdb:pwdevdb@localhost:5432/?sslmode=disable
```

- Auto Migrationが不要な場合は `appuser` で接続

```Shell
export DBURL=postgres://appuser:pwappuser@localhost:5432/devdb?sslmode=disable
```

##### MySQL使用時

```Shell
export DBURL='mysql://devdbuser:pwdevdbuser@tcp(localhost:3306)/devdb?parseTime=true'
```

- Auto Migrationが不要な場合は `appuser` で接続

```Shell
export DBURL='mysql://appuser:pwappuser@tcp(localhost:3306)/devdb?parseTime=true'
```

- 識別用にMySQLの通常の接続文字列の前に `mysql://` を付与

#### アプリケーションの起動

```Shell
go run .
```

### 実行例

#### ユーザを登録

```Shell
curl -X POST -d 'name=user1' -d 'password=pw123' http://localhost:1323/users
```

```JSON
{
  "ID": 1,
  "Name": "user1"
}
```

#### トークンの発行

```Shell
curl -X POST -d 'password=pw123' http://localhost:1323/users/user1/token
```

```JSON
{
  "Token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjc1MTg5NjksImlhdCI6MTcyNzUxMTc2OSwiVXNlcklkIjoxfQ.NHgBt3VlVwukGmQWKVJyGYC3vjcibzQQFsm1apauPac"
}
```

- 有効期限は2時間

#### シェル変数にトークンを設定

```Shell
TOKEN=`curl -s -X POST -d 'password=pw123' http://localhost:1323/users/user1/token | jq -r '.Token'`
```

#### 価格を登録

```Shell
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"DateTime":"2023-05-15 12:10:30", "Store":"pcshop", "Product":"ssd1T", "Price":17800}' http://localhost:1323/v1/prices
```

```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 12:10:30",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17800
}
```

#### 価格の一覧

```Shell
curl -X GET -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/prices
```

```JSON
[
  {
    "ID": 1,
    "DateTime": "2023-05-15 12:10:30",
    "Store": "pcshop",
    "Product": "ssd1T",
    "Price": 17800
  }
]
```

#### 価格の取得

```Shell
curl -X GET -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/prices/1
```

```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 12:10:30",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17800
}
```

#### 価格の更新

```Shell
curl -X PUT -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"DateTime":"2023-05-15 12:10:30", "Store":"pcshop", "Product":"ssd1T", "Price":17500}' http://localhost:1323/v1/prices/1
```

```JSON
{
  "ID": 1,
  "DateTime": "2023-05-15 12:10:30",
  "Store": "pcshop",
  "Product": "ssd1T",
  "Price": 17500
}
```

#### 価格の削除

```Shell
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:1323/v1/prices/1
```

## 開発関連

### テスト

- テスト用に `PostgreSQL` を `15432` ポートで、 `MySQL` を `13306` ポートで起動させた状態でテスト実行
- `PostgreSQL` に接続するテストでは、テスト毎に個別のデータベースを作成し、テストが成功したらデータベースを削除。失敗時は確認用に最後に実行した状態を保持
- `MySQL` に接続するテストでは、テスト毎にユーザ（usersテーブル）を新規登録し、テストが成功しても失敗しても残す。次の実行時のセットアップ処理でusersテーブルからレコード削除

#### データベースのコンテナ起動

```Shell
docker compose up -d --wait postgres-test mysql-test
```

#### テストの実行

##### Testcontainers/Dockertestを使うテストケースを含まない

```Shell
go test -short -count=1 -coverpkg=./handler,./service,./repository ./handler
```

- `make test`でデータベースのコンテナ起動も含めて実行

##### Testcontainers/Dockertestを使うテストケースを含む

```Shell
go test -count=1 -coverpkg=./handler,./service,./repository ./handler
```

- `make testall`でデータベースのコンテナ起動も含めて実行

### Docker

#### イメージの作成

```Shell
docker build -t rest-example .
```

#### データベースのコンテナ起動

##### PostgreSQL

```Shell
docker compose up -d --wait postgres
```

##### MySQL

```Shell
docker compose up -d --wait mysql
```

#### アプリケーションの起動

##### PostgreSQL使用時

```Shell
docker run --init --rm --network=backend-develop -p 1323:1323 -e DBURL=postgres://devdb:pwdevdb@postgres-develop:5432/?sslmode=disable rest-example
```

##### MySQL使用時

```Shell
docker run --init --rm --network=backend-develop -p 1323:1323 -e DBURL='mysql://devdbuser:pwdevdbuser@tcp(mysql-develop:3306)/devdb?parseTime=true' rest-example
```

### 環境変数

| 環境変数名 | 必須 | 説明 |
| ---- | :----: | ---- |
| DBURL | 〇 | データベースの接続文字列<br>例）<br> `postgres://devdb:pwdevdb@localhost:5432/?sslmode=disable` <br> MySQLはプレフィックスに `mysql://` を付与 <br> _※テスト用は固定で変更不可_ |
| JWTKEY |  | 固定したい場合などに任意の文字列を指定。<br>省略した場合、アプリケーションの起動時にランダム生成し、<br>停止すると発行したトークンは有効期限前に **無効** になる |
| ECHOADDRESS |  | 省略時は `:1323` |
