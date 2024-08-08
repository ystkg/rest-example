package handler

import (
	"github.com/ystkg/rest-example/service"
)

func (h *Handler) Service() service.Service {
	return h.service
}

func (h *Handler) SetService(s service.Service) {
	h.service = s
}
