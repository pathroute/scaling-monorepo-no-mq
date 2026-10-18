package platform

import (
	"io"
	"log/slog"
)

// NewLogger returns a JSON logger tagged with the service name.
func NewLogger(w io.Writer, c Config) *slog.Logger {
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})).With("service", c.Service)
}
