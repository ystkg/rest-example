package handler

import (
	"errors"
)

var (
	ErrorNotFound             = errors.New("Not Found")
	ErrorAuthenticationFailed = errors.New("Authentication Failed")
	ErrorAlreadyRegistered    = errors.New("Already Registered")
	ErrorIDCannotRequest      = errors.New("ID cannot be requested")
	ErrorIDUnchangeable       = errors.New("ID is unchangeable")
)
