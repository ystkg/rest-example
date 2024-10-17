package handler

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func NewEcho(h *Handler) *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	e.HTTPErrorHandler = h.errorHandler

	e.Validator = h.validator

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("1K"))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(10))))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         middleware.DefaultSecureConfig.XSSProtection,
		ContentTypeNosniff:    middleware.DefaultSecureConfig.ContentTypeNosniff,
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: `default-uri 'none';`,
		HSTSMaxAge:            31536000,
		HSTSPreloadEnabled:    true,
	}))
	e.Use(noCache)
	e.Use(timeout(h.timeoutSec))
	e.Use(traceRequest)

	e.POST("/users", h.createUser)
	e.POST("/users/:name/token", h.genToken)

	g := e.Group("/v1")
	g.Use(echojwt.WithConfig(h.jwtConfig))

	g.POST("/prices", h.createPrice)
	g.GET("/prices", h.findPrices)
	g.GET("/prices/:id", h.findPrice)
	g.PUT("/prices/:id", h.updatePrice)
	g.DELETE("/prices/:id", h.deletePrice)

	return e
}
