package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func NewSlog() *slog.Logger {
	logger := slog.New(tint.NewHandler(os.Stderr, nil))
	return logger
}
