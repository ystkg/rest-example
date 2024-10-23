package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func somePrices() [][5]any {
	now := time.Now()
	return [][5]any{
		{1, now.Add(-1 * time.Microsecond), "shop1", "memory8G", 3800},
		{1, now, "shop2", "memory16G", 5900},
		{1, now, "shop3", "memory32G", 7600},
		{2, now, "shop3", "ssd1T", 7900},
	}
}

// 価格の登録の正常系
func TestCreatePrice(t *testing.T) {
	testname := "TestCreatePrice"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

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
	rec, diff, _, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 201, rec.Code)

	res := &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	bodyAppendID := fmt.Sprintf(`{"ID":%d, %s`, *res.ID, body[1:])
	assert.JSONEq(t, bodyAppendID, rec.Body.String())

	assert.NotNil(t, diff)
	assert.Equal(t, 1, diff.created.count())
	assert.Zero(t, diff.updated.count())
	assert.Zero(t, diff.logicalDeleted.count())
	assert.Zero(t, diff.physicalDeleted.count())

	assert.Equal(t, 1, len(diff.created.prices))
	entity := diff.created.priceAny() // just one
	assert.Equal(t, *res.ID, entity.ID)
	assert.False(t, entity.DeletedAt.Valid)
	assert.Equal(t, userId, entity.UserID)
	d, _ := time.ParseInLocation(conf.DateTimeLayout, dateTime, conf.Location)
	assert.Equal(t, d, entity.DateTime.In(conf.Location))
	assert.Equal(t, store, entity.Store)
	assert.Equal(t, product, entity.Product)
	assert.Equal(t, price, entity.Price)
}

// 価格の登録のバリデーション
func TestCreatePriceValidation(t *testing.T) {
	testname := "TestCreatePriceValidation"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	jwt := genToken(conf, userId)
	cases := []struct {
		jwt  *string
		body string
		code int
		err  error
	}{
		{nil, "", 401, nil},
		{jwt, "", 400, nil},
		{jwt, "=", 400, nil},
		{jwt, `{"DateTime":"2023-05-15 12:15:30", "Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 201, nil},
		{jwt, `{"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 201, nil},
		{jwt, `{"DateTime":"2023-05-15", "Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 400, nil},
		{jwt, `{"Store":"", "Product":"ssd2T", "Price":1200}`, 400, nil},
		{jwt, `{"Store":"pcshop", "Product":"", "Price":1200}`, 400, nil},
		{jwt, `{"Store":"pcshop", "Product":"ssd2T"}`, 400, nil},
		{jwt, `{"ID":1, "Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 400, handler.ErrIDCannotRequest},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodPost,
			"/v1/prices",
			&v.body,
			echo.MIMEApplicationJSON,
			v.jwt,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}

// 価格の一覧の正常系
func TestFindPrices(t *testing.T) {
	testname := "TestFindPrices"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

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
		genToken(conf, userId),
	)

	// テストの実行
	rec, diff, before, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 200, rec.Code)

	res := &[]api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, diff)

	count := 0
	for _, v := range before.prices {
		if v.UserID == userId && !v.DeletedAt.Valid {
			count++
		}
	}
	assert.Equal(t, count, len(*res))

	for _, v := range *res {
		entity := before.findPrice(*v.ID)
		assert.NotNil(t, entity)
		assert.False(t, entity.DeletedAt.Valid)
		assert.Equal(t, userId, entity.UserID)
		assert.Equal(t, *v.DateTime, entity.DateTime.In(conf.Location).Format(conf.DateTimeLayout))
		assert.Equal(t, v.Store, entity.Store)
		assert.Equal(t, v.Product, entity.Product)
		assert.Equal(t, v.Price, entity.Price)
	}
}

// 価格の一覧のバリデーション
func TestFindPricesValidation(t *testing.T) {
	testname := "TestFindPricesValidation"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	userId := uint(1)
	now := time.Now()
	if _, err := insertPrice(tx, &now, &now, nil, userId, now, "store", "product", 100); err != nil {
		t.Fatal(err)
	}

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	jwt := genToken(conf, userId)
	cases := []struct {
		jwt  *string
		code int
	}{
		{nil, 401},
		{jwt, 200},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodGet,
			"/v1/prices",
			nil,
			"",
			v.jwt,
		)

		// テストの実行
		code, _, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
	}
}

// 価格の取得の正常系
func TestFindPrice(t *testing.T) {
	testname := "TestFindPrice"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	userId := uint(1)
	store, product, price := "pcshop", "ssd1T", uint(9500)
	priceId, err := insertPrice(tx, &now, &now, nil, userId, now, store, product, price)
	if err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	req := newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		genToken(conf, userId),
	)

	// テストの実行
	rec, diff, before, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 200, rec.Code)

	res := &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, diff)

	entity := before.findPrice(priceId)
	assert.False(t, entity.DeletedAt.Valid)
	assert.Equal(t, userId, entity.UserID)
	assert.Equal(t, *res.DateTime, entity.DateTime.In(conf.Location).Format(conf.DateTimeLayout))
	assert.Equal(t, res.Store, entity.Store)
	assert.Equal(t, res.Product, entity.Product)
	assert.Equal(t, res.Price, entity.Price)
}

// 価格の取得のバリデーション
func TestFindPriceValidation(t *testing.T) {
	testname := "TestFindPriceValidation"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	userId := uint(1)
	now := time.Now()
	priceId, err := insertPrice(tx, &now, &now, nil, userId, now, "store", "product", 100)
	if err != nil {
		t.Fatal(err)
	}

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	jwt := genToken(conf, userId)
	priceIdStr := strconv.FormatUint(uint64(priceId), 10)
	cases := []struct {
		jwt     *string
		priceId string
		code    int
		err     error
	}{
		{nil, priceIdStr, 401, nil},
		{jwt, priceIdStr, 200, nil},
		{jwt, priceIdStr + "1", 404, handler.ErrNotFound},
		{jwt, "a", 404, handler.ErrNotFound},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodGet,
			fmt.Sprintf("/v1/prices/%s", v.priceId),
			nil,
			"",
			v.jwt,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}

// 価格の更新の正常系
func TestUpdatePrice(t *testing.T) {
	testname := "TestUpdatePrice"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertPrices(tx, &now, somePrices()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	userId := uint(1)
	priceId := uint(1)
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
	rec, diff, before, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 200, rec.Code)

	res := &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, diff)
	assert.Zero(t, diff.created.count())
	assert.Equal(t, 1, diff.updated.count())
	assert.Zero(t, diff.logicalDeleted.count())
	assert.Zero(t, diff.physicalDeleted.count())

	assert.Equal(t, 1, len(diff.updated.prices))
	entity := diff.updated.priceAny() // just one
	assert.Equal(t, priceId, entity.ID)
	assert.Equal(t, before.findPrice(priceId).CreatedAt, entity.CreatedAt)
	assert.False(t, entity.DeletedAt.Valid)
	assert.Equal(t, userId, entity.UserID)
	d, _ := time.ParseInLocation(conf.DateTimeLayout, *res.DateTime, conf.Location)
	assert.Equal(t, d, entity.DateTime.In(conf.Location))
	assert.Equal(t, res.Store, entity.Store)
	assert.Equal(t, res.Product, entity.Product)
	assert.Equal(t, res.Price, entity.Price)
}

// 価格の更新のバリデーション
func TestUpdatePriceValidation(t *testing.T) {
	testname := "TestUpdatePriceValidation"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	userId := uint(1)
	now := time.Now()
	priceId, err := insertPrice(tx, &now, &now, nil, userId, now, "store", "product", 100)
	if err != nil {
		t.Fatal(err)
	}

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	jwt := genToken(conf, userId)
	priceIdStr := strconv.FormatUint(uint64(priceId), 10)
	cases := []struct {
		jwt     *string
		priceId string
		body    string
		code    int
		err     error
	}{
		{nil, priceIdStr, "", 401, nil},
		{jwt, priceIdStr, "", 400, nil},
		{jwt, priceIdStr, "=", 400, nil},
		{jwt, priceIdStr, `{"DateTime":"2023-05-15 12:15:30", "Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 200, nil},
		{jwt, priceIdStr, `{"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 200, nil},
		{jwt, priceIdStr, fmt.Sprintf(`{"ID":%s ,"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, priceIdStr), 200, nil},
		{jwt, priceIdStr, fmt.Sprintf(`{"ID":%s ,"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, priceIdStr+"1"), 400, handler.ErrIDUnchangeable},
		{jwt, priceIdStr, `{"DateTime":"2023-05-15", "Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 400, nil},
		{jwt, priceIdStr + "1", `{"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 404, handler.ErrNotFound},
		{jwt, "a", `{"Store":"pcshop", "Product":"ssd2T", "Price":1200}`, 404, handler.ErrNotFound},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodPut,
			fmt.Sprintf("/v1/prices/%s", v.priceId),
			&v.body,
			echo.MIMEApplicationJSON,
			v.jwt,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}

// 価格の削除の正常系
func TestDeletePrice(t *testing.T) {
	testname := "TestDeletePrice"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

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
		genToken(conf, userId),
	)

	// テストの実行
	rec, diff, before, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 204, rec.Code)

	assert.NotNil(t, diff)
	assert.Zero(t, diff.created.count())
	assert.Zero(t, diff.updated.count())
	assert.Equal(t, 1, diff.logicalDeleted.count())
	assert.Zero(t, diff.physicalDeleted.count())

	assert.Equal(t, 1, len(diff.logicalDeleted.prices))
	entity := diff.logicalDeleted.priceAny() // just one
	assert.Equal(t, priceId, entity.ID)
	assert.True(t, entity.DeletedAt.Valid)
	beforeEntity := before.findPrice(priceId)
	assert.Equal(t, beforeEntity.CreatedAt, entity.CreatedAt)
	assert.False(t, beforeEntity.DeletedAt.Valid)
	assert.Equal(t, beforeEntity.UserID, entity.UserID)
	assert.Equal(t, beforeEntity.DateTime, entity.DateTime)
	assert.Equal(t, beforeEntity.Store, entity.Store)
	assert.Equal(t, beforeEntity.Product, entity.Product)
	assert.Equal(t, beforeEntity.Price, entity.Price)
}

// 価格の削除のバリデーション
func TestDeletePriceValidation(t *testing.T) {
	testname := "TestDeletePriceValidation"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	userId := uint(1)
	now := time.Now()
	priceId, err := insertPrice(tx, &now, &now, nil, userId, now, "store", "product", 100)
	if err != nil {
		t.Fatal(err)
	}

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	jwt := genToken(conf, userId)
	priceIdStr := strconv.FormatUint(uint64(priceId), 10)
	cases := []struct {
		jwt     *string
		priceId string
		code    int
		err     error
	}{
		{nil, priceIdStr, 401, nil},
		{jwt, priceIdStr + "1", 404, handler.ErrNotFound},
		{jwt, "a", 404, handler.ErrNotFound},
		{jwt, priceIdStr, 204, nil},
		{jwt, priceIdStr, 404, handler.ErrNotFound},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodDelete,
			fmt.Sprintf("/v1/prices/%s", v.priceId),
			nil,
			"",
			v.jwt,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}
