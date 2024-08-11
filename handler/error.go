package handler

import (
	"errors"
	"fmt"
	"net/http"

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

func (h *Handler) customErrorHandler(err error, c echo.Context) {
	var code int
	var detail string
	if httpError, ok := err.(*echo.HTTPError); ok {
		code = httpError.Code
		switch v := httpError.Message.(type) {
		case string:
			detail = v
		case fmt.Stringer:
			detail = v.String()
		case error:
			detail = v.Error()
		}
	}

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

	res := api.ErrorResponse{Title: title}
	if detail != "" {
		res.Detail = &detail
	}

	c.JSONPretty(code, &res, h.indent)
}
