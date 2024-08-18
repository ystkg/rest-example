package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// 価格の登録のリポジトリエラー
func TestCreatePriceCreateError(t *testing.T) {
	testname := "TestCreatePriceCreateError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.price.err, err)
	assert.Nil(t, diff)
}

// 価格の登録のトランザクション開始エラー
func TestCreatePriceBeginTxError(t *testing.T) {
	testname := "TestCreatePriceBeginTxError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.beginTxErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.beginTxErr, err)
	assert.Nil(t, diff)
}

// 価格の登録のコミットエラー
func TestCreatePriceCommitError(t *testing.T) {
	testname := "TestCreatePriceCommitError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.commitErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.commitErr, err)
	assert.Nil(t, diff)
}

// 価格の一覧の検索エラー
func TestFindPricesFindByUserIdError(t *testing.T) {
	testname := "TestFindPricesFindByUserIdError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	req := newRequest(
		http.MethodGet,
		"/v1/prices",
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.price.err, err)
	assert.Nil(t, diff)
}

// 価格の取得の検索エラー
func TestFindPriceFindError(t *testing.T) {
	testname := "TestFindPriceFindError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	store, product, price, inStock := "pcshop", "ssd1T", uint(9500), true
	priceId, err := insertPrice(tx, &now, &now, nil, userId, now, store, product, price, inStock)
	if err != nil {
		t.Fatal(err)
	}
	req := newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.price.err, err)
	assert.Nil(t, diff)
}

// 価格の更新のリポジトリエラー
func TestUpdatePriceUpdateError(t *testing.T) {
	testname := "TestUpdatePriceUpdateError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.price.err, err)
	assert.Nil(t, diff)
}

// 価格の更新の更新件数異常
func TestUpdatePriceRowsAffectedError(t *testing.T) {
	testname := "TestUpdatePriceRowsAffectedError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.rowsAffected, mock.repository.price.overwirte = 2, true

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.repository.price.rowsAffected).Error(), err.Error())
	assert.Nil(t, diff)
}

// 価格の更新のトランザクション開始エラー
func TestUpdatePriceBeginTxError(t *testing.T) {
	testname := "TestUpdatePriceBeginTxError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.beginTxErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.beginTxErr, err)
	assert.Nil(t, diff)
}

// 価格の更新のコミットエラー
func TestUpdatePriceCommitError(t *testing.T) {
	testname := "TestUpdatePriceCommitError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.commitErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
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
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.commitErr, err)
	assert.Nil(t, diff)
}

// 価格の削除のリポジトリエラー
func TestDeletePriceDeleteError(t *testing.T) {
	testname := "TestDeletePriceDeleteError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.price.err, err)
	assert.Nil(t, diff)
}

// 価格の削除の更新件数異常
func TestDeletePriceRowsAffectedError(t *testing.T) {
	testname := "TestDeletePriceRowsAffectedError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.price.rowsAffected, mock.repository.price.overwirte = 2, true

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.repository.price.rowsAffected).Error(), err.Error())
	assert.Nil(t, diff)
}

// 価格の削除のトランザクション開始エラー
func TestDeletePriceBeginTxError(t *testing.T) {
	testname := "TestDeletePriceBeginTxError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.beginTxErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.beginTxErr, err)
	assert.Nil(t, diff)
}

// 価格の削除のコミットエラー
func TestDeletePriceCommitError(t *testing.T) {
	testname := "TestDeletePriceCommitError"

	// セットアップ
	e, sqlDB, mock, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.commitErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.commitErr, err)
	assert.Nil(t, diff)
}
