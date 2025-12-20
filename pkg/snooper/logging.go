package snooper

import (
	"log/slog"
	"os"
)

// SetupLogging configures a default slog logger with the provided level.
func SetupLogging(logLevel slog.Level) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})
	slog.SetDefault(slog.New(h))
}
