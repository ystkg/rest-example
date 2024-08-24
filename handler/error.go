package handler

import (
	"errors"
	"net/http"
	"strings"

	plyerrors "github.com/go-playground/errors/v5"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ystkg/rest-example/api"
)

var (
	// 400
	ErrorAlreadyRegistered = errors.New("already registered")
	ErrorIDCannotRequest   = errors.New("ID cannot be requested")
	ErrorIDUnchangeable    = errors.New("ID is unchangeable")

	// 401
	ErrorAuthenticationFailed = errors.New("authentication failed")

	// 404
	ErrorNotFound = errors.New("not found")
)

func newHTTPError(code int, err error) *echo.HTTPError {
	return echo.NewHTTPError(code, err).SetInternal(plyerrors.WrapSkipFrames(err, "", 1))
}

func (h *Handler) customErrorHandler(err error, c echo.Context) {
	var code int
	var detail string
	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		switch m := he.Message.(type) {
		case string:
			detail = m
		case error:
			detail = m.Error()
		}
	}

	// ステータスコードとエラーメッセージ
	var title string
	switch code {
	case http.StatusBadRequest:
		title = "Parameter Error"
	case http.StatusUnauthorized:
		title = "Authentication Error"
	case http.StatusNotFound:
		title = "Not Found"
	case http.StatusServiceUnavailable:
		title = "System Error"
		detail = "Service Unavailable"
	default:
		code = http.StatusInternalServerError
		title = "System Error"
		detail = "Internal Server Error"
	}

	// バリデーションエラー
	var params []api.InvalidParam
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		trans := verrs.Translate(h.validator.translator)
		params = make([]api.InvalidParam, len(verrs))
		for i, v := range verrs {
			params[i] = api.InvalidParam{
				Name:   v.Field(),
				Reason: trans[v.Namespace()],
			}
		}
	}

	// レスポンスの生成
	res := api.ErrorResponse{
		Title:         title,
		InvalidParams: params,
	}
	if params == nil && detail != "" && !strings.EqualFold(detail, title) {
		res.Detail = &detail
	}
	c.JSONPretty(code, &res, h.indent)

	// エラーログ
	var chain plyerrors.Chain
	if errors.As(err, &chain) {
		h.logger.DebugContext(c.Request().Context(), chain.Error())
	} else {
		h.logger.DebugContext(c.Request().Context(), err.Error())
	}
}
