package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomErrorHandler(t *testing.T) {
	testname := "TestCustomErrorHandler"

	// セットアップ
	h := NewHandler(slog.Default(), nil, nil, 0, nil, "", 0)
	e := NewEcho(h)

	cases := []struct {
		err  error
		code int
	}{
		{newHTTPError(http.StatusBadRequest, ErrorAlreadyRegistered), 400},
		{newHTTPError(http.StatusUnauthorized, ErrorAuthenticationFailed), 401},
		{newHTTPError(http.StatusNotFound, ErrorNotFound), 404},
		{newHTTPError(http.StatusServiceUnavailable, errors.New(testname)), 503},
		{errors.New(testname), 500},
	}

	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// テストの実行
		h.customErrorHandler(v.err, c)

		// アサーション
		assert.Equal(t, v.code, c.Response().Status)
	}
}
