package api

type ErrorResponse struct {
	Title  string  `json:"title"`            // human-readable
	Detail *string `json:"detail,omitempty"` // human-readable
}
