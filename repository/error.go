package repository

import (
	"errors"

	plyerrors "github.com/go-playground/errors/v5"
)

var (
	ErrDuplicated = errors.New("duplicated")
)

func wrap(err error) error {
	return plyerrors.WrapSkipFrames(err, "", 1)
}
