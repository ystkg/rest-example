package handler

import (
	"time"

	"github.com/ystkg/rest-example/service"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service service.Service

	// JWT
	jwtkey      []byte
	validityMin int

	// 日付
	layout   string
	location *time.Location

	// JSON
	indent string

	timeoutSec int
}

func NewHandler(s service.Service, jwtkey []byte, validityMin int, location *time.Location, indent string, timeoutSec int) *Handler {
	return &Handler{
		s,
		jwtkey,
		validityMin,
		time.DateTime,
		location,
		indent,
		timeoutSec,
	}
}

func (h *Handler) newJwtConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &JwtCustomClaims{}
		},
		SigningKey: h.jwtkey,
	}
}

func (h *Handler) userId(c echo.Context) uint {
	return c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims).UserId
}
