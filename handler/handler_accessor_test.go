package handler

import (
	"github.com/ystkg/rest-example/service"
)

// mockに差し替える目的で使う
func (h *Handler) SetMockService(mock service.Service) {
	h.service = mock
}
