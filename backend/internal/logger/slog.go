package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

//func NewSlog() *slog.Logger {
//	logger := slog.New(tint.NewHandler(os.Stderr, nil))
//	return logger
//}

type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger instance
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	if logger == nil {
		logger = slog.New(tint.NewHandler(os.Stderr, nil))
	}
	return &SlogLogger{logger: logger}
}

func (s *SlogLogger) Debug(msg string, args ...interface{}) {
	kvs, _ := argsToKeyValues(args)
	s.logger.Debug(msg, kvs...)
}

func (s *SlogLogger) Info(msg string, args ...interface{}) {
	kvs, _ := argsToKeyValues(args)
	s.logger.Info(msg, kvs...)
}

func (s *SlogLogger) Warn(msg string, args ...interface{}) {
	kvs, _ := argsToKeyValues(args)
	s.logger.Warn(msg, kvs...)
}

func (s *SlogLogger) Error(msg string, args ...interface{}) {
	kvs, _ := argsToKeyValues(args)
	s.logger.Error(msg, kvs...)
}

func (s *SlogLogger) Fatal(msg string, args ...interface{}) {
	kvs, _ := argsToKeyValues(args)
	s.logger.Error(msg, kvs...)
	os.Exit(1)
}

func (s *SlogLogger) With(args ...interface{}) Logger {
	kvs, _ := argsToKeyValues(args)
	return &SlogLogger{
		logger: s.logger.With(kvs...),
	}
}

func (s *SlogLogger) WithContext(ctx context.Context) Logger {
	return &SlogLogger{
		logger: s.logger.With(slog.Any("context", ctx)),
	}
}
