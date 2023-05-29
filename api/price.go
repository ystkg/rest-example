package api

type Price struct {
	ID       *uint
	DateTime *string `validate:"omitempty,max=100"`
	Store    string  `validate:"required,max=100"`
	Product  string  `validate:"required,max=100"`
	Price    uint    `validate:"required"`
	InStock  *bool
}
