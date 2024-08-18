package handler_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ystkg/rest-example/handler"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
)

type anyTime struct{}

func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func setupSqlMockTest(testname string) (*echo.Echo, *sql.DB, sqlmock.Sqlmock, []byte, int, error) {
	logger := slog.Default()

	// Repository
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	r, err := repository.NewRepository(logger, sqlDB)
	if err != nil {
		sqlDB.Close()
		return nil, nil, nil, nil, 0, err
	}

	// Service
	s := service.NewService(logger, r)

	// Handler
	jwtkey := []byte(testname)
	validityMin := 1
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		sqlDB.Close()
		return nil, nil, nil, nil, 0, err
	}
	h := handler.NewHandler(logger, s, &handler.HandlerConfig{
		JwtKey:      jwtkey,
		ValidityMin: validityMin,
		Location:    location,
		Indent:      "  ",
		TimeoutSec:  60,
	})

	// Echo
	e := handler.NewEcho(h)

	return e, sqlDB, mock, jwtkey, validityMin, nil
}

// トランザクション開始エラー
func TestBeginError(t *testing.T) {
	testname := "TestBeginError"

	// セットアップ
	e, sqlDB, mock, _, _, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mockerr := errors.New(testname)
	mock.ExpectBegin().WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act, mockerr)
}

// コミットエラー
func TestCommitError(t *testing.T) {
	testname := "TestCommitError"

	// セットアップ
	e, sqlDB, mock, _, _, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	name, password := "testuser01", "testpassword"
	// PostgreSQLの場合はINSERTでもRETURNがあるのでExpectQueryを使う
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("created_at","updated_at","deleted_at","name","password") `)).
		WithArgs(anyTime{}, anyTime{}, nil, name, encodePassword(name, password)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mockerr := errors.New(testname)
	mock.ExpectCommit().WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act, mockerr)
}

// ユーザの登録のDBエラー
func TestCreateUserError(t *testing.T) {
	testname := "TestCreateUserError"

	// セットアップ
	e, sqlDB, mock, _, _, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	name, password := "testuser01", "testpassword"
	mockerr := errors.New(testname)
	// PostgreSQLの場合はINSERTでもRETURNがあるのでExpectQueryを使う
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("created_at","updated_at","deleted_at","name","password") `)).
		WithArgs(anyTime{}, anyTime{}, nil, name, encodePassword(name, password)).
		WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// トークン発行のDBエラー
func TestGenTokenError(t *testing.T) {
	testname := "TestGenTokenError"

	// セットアップ
	e, sqlDB, mock, _, _, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	name, password := "testuser01", "testpassword"
	limit := 1
	mockerr := errors.New(testname)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "users" `)).
		WithArgs(name, encodePassword(name, password), limit).
		WillReturnError(mockerr)

	// リクエストの生成
	body := "password=" + password
	req := newRequest(
		http.MethodPost,
		"/users/"+name+"/token",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// 価格の登録のDBエラー
func TestCreatePriceError(t *testing.T) {
	testname := "TestCreatePriceError"

	// セットアップ
	e, sqlDB, mock, jwtkey, validityMin, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	mockerr := errors.New(testname)
	// PostgreSQLの場合はINSERTでもRETURNがあるのでExpectQueryを使う
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "prices" ("created_at","updated_at","deleted_at","user_id","date_time","store","product","price","in_stock") `)).
		WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	userId := uint(1)
	dateTime, store, product, price, inStock := "2023-05-19 12:34:56", "pcshop", "ssd1T", uint(9500), true
	body := fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d, "InStock":%t}`, dateTime, store, product, price, inStock)
	req := newRequest(
		http.MethodPost,
		"/v1/prices",
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// 価格の一覧のDBエラー
func TestFindPricesError(t *testing.T) {
	testname := "TestFindPricesError"

	// セットアップ
	e, sqlDB, mock, jwtkey, validityMin, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	userId := uint(1)
	mockerr := errors.New(testname)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "prices" `)).
		WithArgs(userId).
		WillReturnError(mockerr)

	// リクエストの生成
	req := newRequest(
		http.MethodGet,
		"/v1/prices",
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// 価格の取得のDBエラー
func TestFindPriceError(t *testing.T) {
	testname := "TestFindPriceError"

	// セットアップ
	e, sqlDB, mock, jwtkey, validityMin, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	userId := uint(1)
	priceId := uint(2)
	limit := 1
	mockerr := errors.New(testname)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "prices" `)).
		WithArgs(userId, priceId, limit).
		WillReturnError(mockerr)

	// リクエストの生成
	req := newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// 価格の更新のDBエラー
func TestUpdatePriceError(t *testing.T) {
	testname := "TestUpdatePriceError"

	// セットアップ
	e, sqlDB, mock, jwtkey, validityMin, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	mockerr := errors.New(testname)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "prices" SET `)).
		WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	userId := uint(1)
	priceId := uint(2)
	dateTime, store, product, price, inStock := "2023-05-19 12:34:56", "pcshop", "ssd1T", uint(9500), true
	body := fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d, "InStock":%t}`, dateTime, store, product, price, inStock)
	req := newRequest(
		http.MethodPut,
		fmt.Sprintf("/v1/prices/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}

// 価格の削除のDBエラー
func TestDeletePriceError(t *testing.T) {
	testname := "TestDeletePriceError"

	// セットアップ
	e, sqlDB, mock, jwtkey, validityMin, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	userId := uint(1)
	priceId := uint(2)
	mockerr := errors.New(testname)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "prices" SET `)).
		WithArgs(anyTime{}, userId, priceId).
		WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}
