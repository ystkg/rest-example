package handler

import (
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEcho(h *Handler) (*echo.Echo, error) {
	if err := h.InitDB(); err != nil {
		return nil, err
	}

	e := echo.New()
	e.Validator = NewCustomValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/user", h.AddUser)
	e.POST("/user/:name/token", h.GenToken)

	g := e.Group("/v1")
	g.Use(echojwt.WithConfig(h.newJwtConfig()))

	g.POST("/price", h.CreatePrice)
	g.GET("/price", h.FindPrices)
	g.GET("/price/:id", h.FindPrice)
	g.PUT("/price/:id", h.UpdatePrice)
	g.DELETE("/price/:id", h.DeletePrice)

	return e, nil
}
