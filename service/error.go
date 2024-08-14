package service

import (
	"errors"

	pkgerrors "github.com/pkg/errors"
)

var (
	ErrorNotFound = errors.New("not found")
)

func withStack(err error) error {
	return pkgerrors.WithStack(err)
}
