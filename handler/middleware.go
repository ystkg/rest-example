package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/labstack/echo/v4"
)

func timeout(timeoutSec int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), time.Duration(timeoutSec)*time.Second)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func traceRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// トレースIDと見立ててX-Request-Idを使う
		traceID := c.Request().Header.Get(echo.HeaderXRequestID)
		if traceID == "" {
			b := make([]byte, 16)
			rand.Read(b)
			traceID = hex.EncodeToString(b)
			c.Response().Header().Add(echo.HeaderXRequestID, traceID)
		}
		ctx := context.WithValue(c.Request().Context(), contextKeyTraceID, traceID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
