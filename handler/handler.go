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

	validator *customValidator

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

type HandlerConfig struct {
	JwtKey      []byte
	ValidityMin int // JWTのexp
	Location    *time.Location
	Locale      string
	Indent      string // レスポンスのJSONのインデント
	TimeoutSec  int
}

func NewHandler(logger *slog.Logger, s service.Service, config *HandlerConfig) *Handler {
	jwtConfig := echojwt.Config{
		SigningKey: config.JwtKey,
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
		newCustomValidator(config.Locale),
		jwtConfig,
		jwt.GetSigningMethod(signingMethod),
		jwtContextKey,
		config.ValidityMin,
		time.DateTime,
		config.Location,
		config.Indent,
		config.TimeoutSec,
	}
}

func (h *Handler) userId(c echo.Context) uint {
	return c.Get(h.jwtContextKey).(*jwt.Token).Claims.(*JwtCustomClaims).UserId
}
