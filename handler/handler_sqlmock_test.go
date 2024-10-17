package handler_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
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

func setupSqlMockTest(testname string) (*echo.Echo, *handler.HandlerConfig, *sql.DB, sqlmock.Sqlmock, error) {
	// Repository
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	r, err := repository.NewRepository("pgx", sqlDB)
	if err != nil {
		sqlDB.Close()
		return nil, nil, nil, nil, err
	}

	// Service
	s := service.NewService(r)

	// Handler
	jwtkey := []byte(testname)
	validityMin := 1
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		sqlDB.Close()
		return nil, nil, nil, nil, err
	}
	conf := &handler.HandlerConfig{
		JwtKey:           jwtkey,
		ValidityMin:      validityMin,
		DateTimeLayout:   time.DateTime,
		Location:         location,
		Indent:           "  ",
		TimeoutSec:       60,
		RequestBodyLimit: "1K",
		RateLimit:        10,
	}
	h := handler.NewHandler(s, conf)

	// Echo
	e := handler.NewEcho(h)

	return e, conf, sqlDB, mock, nil
}

// SQLドライバエラー
func TestDriverError(t *testing.T) {
	testname := "TestDriverError"

	// セットアップ
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	// テストの実行
	r, act := repository.NewRepository(testname, sqlDB)

	// アサーション
	assert.Nil(t, r)
	assert.Equal(t, fmt.Errorf("unsupported:%s", testname), act)
}

// Pingエラー
func TestPingError(t *testing.T) {
	testname := "TestPingError"

	// セットアップ
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}

	// mockの挙動設定
	mockerr := errors.New(testname)
	mock.ExpectPing().WillReturnError(mockerr)

	// テストの実行
	r, act := repository.NewRepository("pgx", sqlDB)

	// アサーション
	assert.Nil(t, r)
	assert.ErrorIs(t, act, mockerr)
}

// トランザクション開始エラー
func TestBeginError(t *testing.T) {
	testname := "TestBeginError"

	// セットアップ
	e, _, sqlDB, mock, err := setupSqlMockTest(testname)
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
	e, _, sqlDB, mock, err := setupSqlMockTest(testname)
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
	e, _, sqlDB, mock, err := setupSqlMockTest(testname)
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
	e, _, sqlDB, mock, err := setupSqlMockTest(testname)
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
	e, conf, sqlDB, mock, err := setupSqlMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// mockの挙動設定
	mock.ExpectBegin()
	mockerr := errors.New(testname)
	// PostgreSQLの場合はINSERTでもRETURNがあるのでExpectQueryを使う
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "prices" ("created_at","updated_at","deleted_at","user_id","date_time","store","product","price") `)).
		WillReturnError(mockerr)
	mock.ExpectRollback()

	// リクエストの生成
	userId := uint(1)
	dateTime, store, product, price := "2023-05-19 12:34:56", "pcshop", "ssd1T", uint(9500)
	body := fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d}`, dateTime, store, product, price)
	req := newRequest(
		http.MethodPost,
		"/v1/prices",
		&body,
		echo.MIMEApplicationJSON,
		genToken(conf, userId),
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
	e, conf, sqlDB, mock, err := setupSqlMockTest(testname)
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
		genToken(conf, userId),
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
	e, conf, sqlDB, mock, err := setupSqlMockTest(testname)
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
		genToken(conf, userId),
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
	e, conf, sqlDB, mock, err := setupSqlMockTest(testname)
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
	dateTime, store, product, price := "2023-05-19 12:34:56", "pcshop", "ssd1T", uint(9500)
	body := fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d}`, dateTime, store, product, price)
	req := newRequest(
		http.MethodPut,
		fmt.Sprintf("/v1/prices/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(conf, userId),
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
	e, conf, sqlDB, mock, err := setupSqlMockTest(testname)
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
		genToken(conf, userId),
	)

	// テストの実行
	_, act := execHandler(e, req)

	// アサーション
	assert.ErrorIs(t, act.(interface{ Unwrap() error }).Unwrap(), mockerr)
}
