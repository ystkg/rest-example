package handler

import (
	"log/slog"
	"time"

	"github.com/ystkg/rest-example/service"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	logger *slog.Logger

	service service.Service

	// JWT
	jwtConfig     echojwt.Config
	signingMethod jwt.SigningMethod
	jwtContextKey string
	validityMin   int

	// 日付
	layout   string
	location *time.Location

	// JSON
	indent string

	timeoutSec int
}

func NewHandler(logger *slog.Logger, s service.Service, jwtkey []byte, validityMin int, location *time.Location, indent string, timeoutSec int) *Handler {
	jwtConfig := echojwt.Config{
		SigningKey: jwtkey,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &JwtCustomClaims{}
		},
	}
	signingMethod := jwtConfig.SigningMethod
	if signingMethod == "" {
		signingMethod = echojwt.AlgorithmHS256 // デフォルトはHS256になる
	}
	jwtContextKey := jwtConfig.ContextKey
	if jwtContextKey == "" {
		jwtContextKey = "user" // デフォルトは"user"になる
	}

	return &Handler{
		logger,
		s,
		jwtConfig,
		jwt.GetSigningMethod(signingMethod),
		jwtContextKey,
		validityMin,
		time.DateTime,
		location,
		indent,
		timeoutSec,
	}
}

func (h *Handler) userId(c echo.Context) uint {
	return c.Get(h.jwtContextKey).(*jwt.Token).Claims.(*JwtCustomClaims).UserId
}
