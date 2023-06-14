package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (h *Handler) AddUser(c echo.Context) error {
	req := &api.User{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.ID != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorIDCannotRequest.Error())
	}

	user := &entity.User{
		Name:     req.Name,
		Password: encodePassword(req.Name, req.Password),
	}

	tx, err := h.beginTX()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return echo.NewHTTPError(http.StatusBadRequest, ErrorAlreadyRegistered.Error())
		}
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			return echo.NewHTTPError(http.StatusBadRequest, ErrorAlreadyRegistered.Error())
		}
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return c.JSONPretty(http.StatusCreated, &api.User{ID: &user.ID, Name: user.Name}, "  ")
}

func (h *Handler) GenToken(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}
	password := encodePassword(name, c.FormValue("password"))

	user := &entity.User{}
	if db := h.db.Select("id").Where("name = ? AND password = ?", name, password).First(user); db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusForbidden, ErrorAuthenticationFailed.Error())
		}
		return db.Error
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&JwtCustomClaims{
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)),
			},
			user.ID,
		},
	)
	signed, err := token.SignedString(h.jwtkey)
	if err != nil {
		return err
	}

	return c.JSONPretty(http.StatusCreated, &api.UserToken{Token: signed}, "  ")
}

func encodePassword(user, password string) string {
	sha256 := sha256.Sum256([]byte(fmt.Sprintf("%s %s", user, password)))
	return hex.EncodeToString(sha256[:])
}
