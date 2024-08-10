package handler

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEcho(h *Handler) *echo.Echo {
	e := echo.New()

	e.Validator = NewCustomValidator()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	root := e.Group("/")
	root.Use(timeout(h.timeoutSec))

	root.POST("user", h.CreateUser)
	root.POST("user/:name/token", h.GenToken)

	g := root.Group("v1/")
	g.Use(echojwt.WithConfig(h.jwtConfig))

	g.POST("price", h.CreatePrice)
	g.GET("price", h.FindPrices)
	g.GET("price/:id", h.FindPrice)
	g.PUT("price/:id", h.UpdatePrice)
	g.DELETE("price/:id", h.DeletePrice)

	return e
}
