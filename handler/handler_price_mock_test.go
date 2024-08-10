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

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestCreatePriceBeginTxError(t *testing.T) {
	testname := "TestCreatePriceBeginTxError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestCreatePriceCommitError(t *testing.T) {
	testname := "TestCreatePriceCommitError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}

func TestFindPricesFindByUserIdError(t *testing.T) {
	testname := "TestFindPricesFindByUserIdError"

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

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)

	req := newRequest(
		http.MethodGet,
		"/v1/price",
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestFindPriceFindError(t *testing.T) {
	testname := "TestFindPriceFindError"

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

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceUpdateError(t *testing.T) {
	testname := "TestUpdatePriceUpdateError"

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

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceRowsAffectedError(t *testing.T) {
	testname := "TestUpdatePriceRowsAffectedError"

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

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.price.(*priceRepositoryMock).rowsAffected), err)
	assert.Nil(t, diff)
}

func TestUpdatePriceBeginTxError(t *testing.T) {
	testname := "TestUpdatePriceBeginTxError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestUpdatePriceCommitError(t *testing.T) {
	testname := "TestUpdatePriceCommitError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}

func TestDeletePriceDeleteError(t *testing.T) {
	testname := "TestDeletePriceDeleteError"

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

	now := time.Now()

	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	priceId := uint(1)

	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.price.(*priceRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestDeletePriceRowsAffectedError(t *testing.T) {
	testname := "TestDeletePriceRowsAffectedError"

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

	now := time.Now()

	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	priceId := uint(1)

	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, fmt.Errorf("RowsAffected:%d", mock.price.(*priceRepositoryMock).rowsAffected), err)
	assert.Nil(t, diff)
}

func TestDeletePriceBeginTxError(t *testing.T) {
	testname := "TestDeletePriceBeginTxError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()

	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	priceId := uint(1)

	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestDeletePriceCommitError(t *testing.T) {
	testname := "TestDeletePriceCommitError"

	e, h, tx, jwtkey, validityMin, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetMockService(newMockService(mock))

	now := time.Now()

	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	priceId := uint(1)

	req := newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/price/%d", priceId),
		nil,
		"",
		genToken(userId, jwtkey, validityMin),
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}
