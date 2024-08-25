package handler

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	contextKeyTraceID = contextKey("traceID")
)

type customLogHandler struct {
	slog.Handler
}

func (h *customLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID := ctx.Value(contextKeyTraceID); traceID != nil {
		r.Add("traceID", traceID)
	}

	return h.Handler.Handle(ctx, r)
}

func NewLogger() *slog.Logger {
	return slog.New(&customLogHandler{slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	)})
}
