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

func TestMiddlewareTimeout(t *testing.T) {
	testname := "TestMiddlewareTimeout"

	// セットアップ
	const timeoutSec = 60
	h := NewHandler(nil, &HandlerConfig{TimeoutSec: timeoutSec, RequestBodyLimit: "1K"})
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

func TestTraceRequestWithReqHeader(t *testing.T) {
	testname := "TestTraceRequestWithReqHeader"

	// セットアップ
	e := NewEcho(NewHandler(nil, &HandlerConfig{RequestBodyLimit: "1K"}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	const traceID = "tid"
	req.Header.Add(echo.HeaderXRequestID, traceID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := errors.New(testname)
	called := false
	next := func(c echo.Context) error {
		called = true
		return err
	}

	// テストの実行
	ret := traceRequest(next)(c)

	// アサーション
	assert.Equal(t, traceID, c.Request().Context().Value(contextKeyTraceID))
	assert.Empty(t, rec.Header().Get(echo.HeaderXRequestID))
	assert.True(t, called)
	assert.Equal(t, ret, err)
}

func TestTraceRequestWithoutReqHeader(t *testing.T) {
	testname := "TestTraceRequestWithoutReqHeader"

	// セットアップ
	e := NewEcho(NewHandler(nil, &HandlerConfig{RequestBodyLimit: "1K"}))

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
	ret := traceRequest(next)(c)

	// アサーション
	traceID := rec.Header().Get(echo.HeaderXRequestID)
	assert.NotEmpty(t, traceID)
	assert.Equal(t, traceID, c.Request().Context().Value(contextKeyTraceID))
	assert.True(t, called)
	assert.Equal(t, ret, err)
}

func TestNoCache(t *testing.T) {
	testname := "TestNoCache"

	// セットアップ
	e := NewEcho(NewHandler(nil, &HandlerConfig{RequestBodyLimit: "1K"}))

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
	ret := noCache(next)(c)

	// アサーション
	assert.Equal(t, "no-store", rec.Header().Get(echo.HeaderCacheControl))
	assert.True(t, called)
	assert.Equal(t, ret, err)
}
