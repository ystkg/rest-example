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

	e.POST("/users", h.CreateUser)
	e.POST("/users/:name/token", h.GenToken)

	g := e.Group("/v1")
	g.Use(echojwt.WithConfig(h.jwtConfig))

	g.POST("/prices", h.CreatePrice)
	g.GET("/prices", h.FindPrices)
	g.GET("/prices/:id", h.FindPrice)
	g.PUT("/prices/:id", h.UpdatePrice)
	g.DELETE("/prices/:id", h.DeletePrice)

	return e
}
