package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ystkg/rest-example/api"
)

// シナリオテスト（MySQL使用）
func TestScenario(t *testing.T) {
	testname := "TestScenario"

	// セットアップ
	e, err := setupMySQLTest(testname)
	if err != nil {
		t.Fatal(err)
	}

	// テストシナリオの実行
	scenario(t, e, testname)
}

// テストシナリオ
func scenario(t *testing.T, e *echo.Echo, testname string) {
	// ユーザの登録
	name, password := testname, "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)
	rec, err := execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 201, rec.Code)

	// ユーザの登録（重複）
	rec, err = execHandler(e, req)
	assert.NotNil(t, err)
	httperr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 400, httperr.Code)

	// トークン発行
	body = "password=" + password
	req = newRequest(
		http.MethodPost,
		"/users/"+name+"/token",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 201, rec.Code)
	resToken := &api.UserToken{}
	if err := json.Unmarshal(rec.Body.Bytes(), resToken); err != nil {
		t.Fatal(err)
	}
	token := resToken.Token

	// 価格の登録
	dateTime, store, product, price := "2024-09-28 12:34:56", "pcshop", "ssd1T", uint(9500)
	body = fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d}`, dateTime, store, product, price)
	req = newRequest(
		http.MethodPost,
		"/v1/prices",
		&body,
		echo.MIMEApplicationJSON,
		&token,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 201, rec.Code)
	resPrice := &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), resPrice); err != nil {
		t.Fatal(err)
	}
	priceId := *resPrice.ID

	// 価格の取得
	req = newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		&token,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, rec.Code)
	resPrice = &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), resPrice); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, priceId, *resPrice.ID)
	assert.Equal(t, dateTime, *resPrice.DateTime)
	assert.Equal(t, store, resPrice.Store)
	assert.Equal(t, product, resPrice.Product)
	assert.Equal(t, price, resPrice.Price)

	// 価格の更新
	dateTime, store, product, price = "2024-09-28 23:45:01", "pcstore", "ssd2T", uint(9200)
	body = fmt.Sprintf(`{"DateTime":"%s", "Store":"%s", "Product":"%s", "Price":%d}`, dateTime, store, product, price)
	req = newRequest(
		http.MethodPut,
		fmt.Sprintf("/v1/prices/%d", priceId),
		&body,
		echo.MIMEApplicationJSON,
		&token,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, rec.Code)

	// 価格の取得
	req = newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		&token,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, rec.Code)
	resPrice = &api.Price{}
	if err := json.Unmarshal(rec.Body.Bytes(), resPrice); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, priceId, *resPrice.ID)
	assert.Equal(t, dateTime, *resPrice.DateTime)
	assert.Equal(t, store, resPrice.Store)
	assert.Equal(t, product, resPrice.Product)
	assert.Equal(t, price, resPrice.Price)

	// 価格の削除
	req = newRequest(
		http.MethodDelete,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		&token,
	)
	rec, err = execHandler(e, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 204, rec.Code)

	// 価格の取得
	req = newRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/prices/%d", priceId),
		nil,
		"",
		&token,
	)
	rec, err = execHandler(e, req)
	assert.NotNil(t, err)
	httperr, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, 404, httperr.Code)
}
