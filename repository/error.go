package repository

import (
	"errors"

	pkgerrors "github.com/pkg/errors"
)

var (
	ErrorDuplicated = errors.New("duplicated")
)

func withStack(err error) error {
	return pkgerrors.WithStack(err)
}
