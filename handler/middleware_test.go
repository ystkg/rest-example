package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareTimeout(t *testing.T) {
	testname := "TestMiddlewareTimeout"

	// セットアップ
	const timeoutSec = 60
	h := NewHandler(slog.Default(), nil, nil, 0, nil, "", timeoutSec)
	e := NewEcho(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := errors.New(testname)
	called := false
	next := func(c echo.Context) error {
		called = true
		return err
	}

	// テストの実行
	before := time.Now()
	ret := timeout(h.timeoutSec)(next)(c)
	after := time.Now()

	// アサーション
	act, ok := c.Request().Context().Deadline()
	assert.True(t, ok)
	assert.LessOrEqual(t, before.Add(time.Duration(h.timeoutSec)*time.Second), act)
	assert.GreaterOrEqual(t, after.Add(time.Duration(h.timeoutSec)*time.Second), act)

	assert.True(t, called)
	assert.Equal(t, ret, err)
}
