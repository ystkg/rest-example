package api

type ErrorResponse struct {
	Title         string         `json:"title"`            // human-readable
	Detail        *string        `json:"detail,omitempty"` // human-readable
	InvalidParams []InvalidParam `json:"invalid-params,omitempty"`
}

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"` // human-readable
}
