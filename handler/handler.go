package handler

import (
	"time"

	"github.com/ystkg/rest-example/entity"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db       *gorm.DB
	jwtkey   []byte
	location *time.Location
	layout   string
}

func NewHandler(db *gorm.DB, jwtkey []byte, location *time.Location) *Handler {
	return &Handler{
		db,
		jwtkey,
		location,
		time.DateTime,
	}
}

func (h *Handler) InitDB() error {
	return h.db.AutoMigrate(
		&entity.User{},
		&entity.Price{},
	)
}

func (h *Handler) beginTX() (*gorm.DB, error) {
	tx := h.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	return tx, nil
}

func (h *Handler) newJwtConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return JwtCustomClaims{}
		},
		SigningKey: h.jwtkey,
	}
}
