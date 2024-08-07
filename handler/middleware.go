package handler

import (
	"context"
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
