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

	// Limit
	requestBodyLimit string
	rateLimit        int
}

type HandlerConfig struct {
	JwtKey           []byte
	ValidityMin      int // JWTのexp
	DateTimeLayout   string
	Location         *time.Location
	Locale           string
	Indent           string // レスポンスのJSONのインデント
	TimeoutSec       int
	RequestBodyLimit string
	RateLimit        int
}

func NewHandler(s service.Service, config *HandlerConfig) *Handler {
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
		service:          s,
		validator:        newValidator(config.Locale),
		jwtConfig:        jwtConfig,
		signingMethod:    jwt.GetSigningMethod(signingMethod),
		jwtContextKey:    jwtContextKey,
		validityMin:      config.ValidityMin,
		layout:           config.DateTimeLayout,
		location:         config.Location,
		indent:           config.Indent,
		timeoutSec:       config.TimeoutSec,
		requestBodyLimit: config.RequestBodyLimit,
		rateLimit:        config.RateLimit,
	}
}

func (h *Handler) userId(c echo.Context) uint {
	return c.Get(h.jwtContextKey).(*jwt.Token).Claims.(*JwtCustomClaims).UserId
}
