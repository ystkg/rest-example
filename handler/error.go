package handler

import (
	"errors"
)

var (
	ErrorNotFound             = errors.New("not found")
	ErrorAuthenticationFailed = errors.New("authentication failed")
	ErrorAlreadyRegistered    = errors.New("already registered")
	ErrorIDCannotRequest      = errors.New("ID cannot be requested")
	ErrorIDUnchangeable       = errors.New("ID is unchangeable")
)
