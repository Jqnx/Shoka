// Package log builds the application's structured logger, routing zerolog
// output through a slog handler.
package log

import (
	"Shoka/internal/config"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New returns a slog logger writing console-formatted records to stdout at
// the level configured in cfg, falling back to info when that level is
// unparseable.
func New(cfg *config.Config) *slog.Logger {
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zl := zerolog.New(output).Level(logLevel).With().Timestamp().Logger()
	handler := zerolog.NewSlogHandler(zl)
	logger := slog.New(handler)

	return logger
}
