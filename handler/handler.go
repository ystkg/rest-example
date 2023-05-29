package handler

import (
	"time"

	"github.com/ystkg/rest-example/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	dburl    string
	jwtkey   []byte
	location *time.Location
	layout   string
}

func NewHandler(dburl string, jwtkey []byte, location *time.Location) *Handler {
	return &Handler{
		dburl,
		jwtkey,
		location,
		time.DateTime,
	}
}

func (h *Handler) InitDB() error {
	db, err := h.openDB()
	if err != nil {
		return err
	}
	return db.AutoMigrate(
		&entity.User{},
		&entity.Price{},
	)
}

func (h *Handler) openDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(h.dburl), &gorm.Config{})
}

func (h *Handler) beginTX() (*gorm.DB, error) {
	db, err := h.openDB()
	if err != nil {
		return nil, err
	}
	tx := db.Begin()
	if tx.Error != nil {
		err := tx.Error
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}

func (h *Handler) newJwtConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JwtCustomClaims)
		},
		SigningKey: h.jwtkey,
	}
}
