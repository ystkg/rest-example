package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
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

func newHTTPError(code int, err error) *echo.HTTPError {
	return echo.NewHTTPError(code).SetInternal(pkgerrors.WithStack(err))
}

func (h *Handler) customErrorHandler(err error, c echo.Context) {
	var code int
	var detail string
	var internal error
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		internal = he.Internal
		if _, ok := internal.(interface{ Unwrap() error }); ok {
			internal = internal.(interface{ Unwrap() error }).Unwrap()
			if herr, ok := internal.(*echo.HTTPError); ok && herr.Internal != nil {
				internal = herr.Internal
			}
		}
		switch m := he.Message.(type) {
		case string:
			detail = m
		case error:
			detail = m.Error()
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

	var params []api.InvalidParam
	if verrs, ok := internal.(validator.ValidationErrors); ok {
		trans := verrs.Translate(h.validator.translator)
		params = make([]api.InvalidParam, len(verrs))
		for i, v := range verrs {
			params[i] = api.InvalidParam{
				Name:   v.Field(),
				Reason: trans[v.Namespace()],
			}
		}
	}

	res := api.ErrorResponse{
		Title:         title,
		InvalidParams: params,
	}
	if params == nil && detail != "" {
		res.Detail = &detail
	}

	c.JSONPretty(code, &res, h.indent)

	for errs, causes := []error{err}, []error{}; 0 < len(errs); errs, causes = causes, []error{} {
		for _, v := range errs {
			h.logger.DebugContext(c.Request().Context(), fmt.Sprintf("%+v", v))

			if e, ok := v.(interface{ Unwrap() []error }); ok {
				causes = append(causes, e.Unwrap()...)
			} else if e, ok := v.(interface{ Unwrap() error }); ok {
				causes = append(causes, e.Unwrap())
			}
		}
	}
}
