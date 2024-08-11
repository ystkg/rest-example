package handler

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEcho(h *Handler) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = h.customErrorHandler

	e.Validator = NewCustomValidator()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(timeout(h.timeoutSec))

	e.POST("/user", h.CreateUser)
	e.POST("/user/:name/token", h.GenToken)

	g := e.Group("/v1")
	g.Use(echojwt.WithConfig(h.jwtConfig))

	g.POST("/price", h.CreatePrice)
	g.GET("/price", h.FindPrices)
	g.GET("/price/:id", h.FindPrice)
	g.PUT("/price/:id", h.UpdatePrice)
	g.DELETE("/price/:id", h.DeletePrice)

	return e
}
