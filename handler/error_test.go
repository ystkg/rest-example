package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/handler"
)

func TestErrorHandler(t *testing.T) {
	testname := "TestErrorHandler"

	// セットアップ
	h := handler.NewHandler(nil, &handler.HandlerConfig{RequestBodyLimit: "1K"})
	e := handler.NewEcho(h)

	cause := errors.New(testname)
	cases := []struct {
		err  error
		code int
	}{
		{echo.NewHTTPError(http.StatusBadRequest).SetInternal(cause), 400},
		{echo.NewHTTPError(http.StatusBadRequest, testname).SetInternal(cause), 400},
		{echo.NewHTTPError(http.StatusBadRequest, cause).SetInternal(cause), 400},
		{echo.NewHTTPError(http.StatusUnauthorized).SetInternal(cause), 401},
		{echo.NewHTTPError(http.StatusNotFound).SetInternal(cause), 404},
		{echo.NewHTTPError(http.StatusRequestEntityTooLarge).SetInternal(cause), 413},
		{echo.NewHTTPError(http.StatusTooManyRequests).SetInternal(cause), 429},
		{echo.NewHTTPError(http.StatusServiceUnavailable).SetInternal(cause), 503},
		{errors.Join(cause), 500},
	}

	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// テストの実行
		e.HTTPErrorHandler(v.err, c)

		// アサーション
		assert.Equal(t, v.code, c.Response().Status)
	}
}

func TestErrorHandlerEn(t *testing.T) {
	// セットアップ
	h := handler.NewHandler(nil, &handler.HandlerConfig{RequestBodyLimit: "1K"})
	e := handler.NewEcho(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	param := &api.User{Name: "user1"}
	he := echo.NewHTTPError(http.StatusBadRequest).SetInternal(c.Validate(param))

	// テストの実行
	e.HTTPErrorHandler(he, c)

	// アサーション
	assert.Equal(t, 400, c.Response().Status)

	var rerRes api.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &rerRes); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(rerRes.InvalidParams))
	assert.Equal(t, "Password", rerRes.InvalidParams[0].Name)
	assert.Equal(t, "Password is a required field", rerRes.InvalidParams[0].Reason)
}

func TestErrorHandlerJa(t *testing.T) {
	// セットアップ
	h := handler.NewHandler(nil, &handler.HandlerConfig{Locale: "ja", RequestBodyLimit: "1K"})
	e := handler.NewEcho(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	param := &api.User{Name: "user1"}
	he := echo.NewHTTPError(http.StatusBadRequest).SetInternal(c.Validate(param))

	// テストの実行
	e.HTTPErrorHandler(he, c)

	// アサーション
	assert.Equal(t, 400, c.Response().Status)

	var rerRes api.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &rerRes); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(rerRes.InvalidParams))
	assert.Equal(t, "Password", rerRes.InvalidParams[0].Name)
	assert.Equal(t, "Passwordは必須フィールドです", rerRes.InvalidParams[0].Reason)
}
