package api

import "log/slog"

type password string

func (p password) masked() string {
	return "*"
}

func (p password) String() string {
	return p.masked()
}

func (p password) GoString() string {
	return p.masked()
}

func (p password) LogValue() slog.Value {
	return slog.StringValue(p.masked())
}

type User struct {
	ID       *uint
	Name     string   `form:"name" validate:"required,alphanum,max=30"`
	Password password `form:"password" validate:"required,printascii,max=50" json:"-"`
}

type UserToken struct {
	Token string
}
