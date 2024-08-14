package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	pkgerrors "github.com/pkg/errors"
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

func newHTTPError(code int, cause error) *echo.HTTPError {
	err := echo.NewHTTPError(code)
	err.SetInternal(pkgerrors.WithStack(cause))
	return err
}

func (h *Handler) customErrorHandler(err error, c echo.Context) {
	var code int
	var detail string
	if httpError, ok := err.(*echo.HTTPError); ok {
		code = httpError.Code
		detail = httpError.Internal.(interface{ Unwrap() error }).Unwrap().Error()
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

	errs := []error{err}
	for 0 < len(errs) {
		causes := []error{}
		for _, v := range errs {
			h.logger.DebugContext(c.Request().Context(), fmt.Sprintf("%+v", v))
			if e, ok := v.(interface{ Unwrap() []error }); ok {
				causes = append(causes, e.Unwrap()...)
			} else if e, ok := v.(interface{ Unwrap() error }); ok {
				causes = append(causes, e.Unwrap())
			}
		}
		errs = causes
	}
}
