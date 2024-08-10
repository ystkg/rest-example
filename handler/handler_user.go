package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// ユーザの登録
func (h *Handler) CreateUser(c echo.Context) error {
	// リクエストの取得
	req := &api.User{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// 入力チェック
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// サービスの実行
	userId, err := h.service.CreateUser(c.Request().Context(), req.Name, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrorDuplicated) {
			return echo.NewHTTPError(http.StatusBadRequest, ErrorAlreadyRegistered.Error())
		}
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusCreated, &api.User{ID: userId, Name: req.Name}, h.indent)
}

// トークン発行
func (h *Handler) GenToken(c echo.Context) error {
	// リクエストの取得
	name := c.Param("name")
	password := c.FormValue("password")

	// 入力チェック
	if name == "" {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	// サービスの実行
	userId, err := h.service.FindUser(c.Request().Context(), name, password)
	if err != nil {
		return err
	}
	if userId == nil {
		return echo.NewHTTPError(http.StatusForbidden, ErrorAuthenticationFailed.Error())
	}

	// トークンの生成
	iat := time.Now()
	token := jwt.NewWithClaims(
		h.signingMethod,
		&JwtCustomClaims{
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(iat.Add(time.Duration(h.validityMin) * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(iat),
			},
			*userId,
		},
	)
	signed, err := token.SignedString(h.jwtConfig.SigningKey)
	if err != nil {
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusCreated, &api.UserToken{Token: signed}, h.indent)
}
