package handler

import (
	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	jwt.RegisteredClaims

	UserId uint
}
