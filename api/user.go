package api

type User struct {
	ID       *uint
	Name     string `form:"name" validate:"required,alphanum,max=30"`
	Password string `form:"password" validate:"required,printascii,max=50" json:"-"`
}

type UserToken struct {
	Token string
}
