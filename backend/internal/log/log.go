package log

import (
	"Shoka/internal/config"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(cfg *config.Config) *slog.Logger {
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zl := zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()
	handler := zerolog.NewSlogHandler(zl)
	logger := slog.New(handler)

	return logger
}
