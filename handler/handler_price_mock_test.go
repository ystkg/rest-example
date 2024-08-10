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

func TestCreatePriceCreateError(t *testing.T) {
	testname := "TestCreatePriceCreateError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).err = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		"/v1/price",
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestCreatePriceBeginTxError(t *testing.T) {
	testname := "TestCreatePriceBeginTxError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		"/v1/price",
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestCreatePriceCommitError(t *testing.T) {
	testname := "TestCreatePriceCommitError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		"/v1/price",
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}

func TestFindPricesFindByUserIdError(t *testing.T) {
	testname := "TestFindPricesFindByUserIdError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).err = errors.New(testname)
	h.SetMockService(newMockService(mock))

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	req := newRequest(
		http.MethodGet,
		"/v1/price",
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestFindPriceFindError(t *testing.T) {
	testname := "TestFindPriceFindError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).err = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceUpdateError(t *testing.T) {
	testname := "TestUpdatePriceUpdateError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).err = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceRowsAffectedError(t *testing.T) {
	testname := "TestUpdatePriceRowsAffectedError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).rowsAffected, mock.price.(*priceRepositoryMock).overwirte = 2, true
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.price.(*priceRepositoryMock).rowsAffected), err)
	assert.Nil(t, diff)
}

func TestUpdatePriceBeginTxError(t *testing.T) {
	testname := "TestUpdatePriceBeginTxError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceCommitError(t *testing.T) {
	testname := "TestUpdatePriceCommitError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}

func TestDeletePriceDeleteError(t *testing.T) {
	testname := "TestDeletePriceDeleteError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).err = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestDeletePriceRowsAffectedError(t *testing.T) {
	testname := "TestDeletePriceRowsAffectedError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.price = newMockPriceRepository(h.Service().(*serviceMock))
	mock.price.(*priceRepositoryMock).rowsAffected, mock.price.(*priceRepositoryMock).overwirte = 2, true
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.price.(*priceRepositoryMock).rowsAffected), err)
	assert.Nil(t, diff)
}

func TestDeletePriceBeginTxError(t *testing.T) {
	testname := "TestDeletePriceBeginTxError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestDeletePriceCommitError(t *testing.T) {
	testname := "TestDeletePriceCommitError"

	// セットアップ
	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

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
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}
