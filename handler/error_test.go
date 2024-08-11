package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCustomErrorHandler(t *testing.T) {
	testname := "TestCustomErrorHandler"

	// セットアップ
	h := NewHandler(nil, nil, 0, nil, "", 0)
	e := NewEcho(h)

	cases := []struct {
		err  error
		code int
	}{
		{echo.NewHTTPError(http.StatusBadRequest, ErrorAlreadyRegistered), 400},
		{echo.NewHTTPError(http.StatusUnauthorized, ErrorAuthenticationFailed), 401},
		{echo.NewHTTPError(http.StatusNotFound, ErrorNotFound), 404},
		{echo.NewHTTPError(http.StatusServiceUnavailable, testname), 503},
		{echo.NewHTTPError(http.StatusServiceUnavailable, time.Now()), 503}, // fmt.Stringer
		{errors.New(testname), 500},
	}

	for _, v := range cases {
		rec := httptest.NewRecorder()
		c := e.NewContext(nil, rec)

		// テストの実行
		h.customErrorHandler(v.err, c)

		// アサーション
		assert.Equal(t, v.code, c.Response().Status)
	}
}
