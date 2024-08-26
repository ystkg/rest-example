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
	h := NewHandler(nil, &HandlerConfig{TimeoutSec: timeoutSec})
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
	assert.LessOrEqual(t, before.UTC().Add(time.Duration(h.timeoutSec)*time.Second), act.UTC())
	assert.GreaterOrEqual(t, after.UTC().Add(time.Duration(h.timeoutSec)*time.Second), act.UTC())

	assert.True(t, called)
	assert.Equal(t, ret, err)
}

func TestTraceRequestWithReqHeader(t *testing.T) {
	testname := "TestTraceRequestWithReqHeader"

	// セットアップ
	e := NewEcho(NewHandler(nil, &HandlerConfig{}))

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
	e := NewEcho(NewHandler(nil, &HandlerConfig{}))

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
