package service

import (
	"errors"

	plyerrors "github.com/go-playground/errors/v5"
)

var (
	ErrorNotFound = errors.New("not found")
)

func wrap(err error) error {
	return plyerrors.WrapSkipFrames(err, "", 1)
}
