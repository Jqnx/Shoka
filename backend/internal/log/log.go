package log

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func New() *slog.Logger {
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = int(zerolog.InfoLevel)
	}

	zl := zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()
	handler := zerolog.NewSlogHandler(zl)
	logger := slog.New(handler)

	return logger
}
