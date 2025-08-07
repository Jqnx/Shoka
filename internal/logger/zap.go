package logger

import (
	"context"

	"go.uber.org/zap"
)

// ZapLogger wraps zap.Logger
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewZapLogger creates a new ZapLogger instance
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

func (z *ZapLogger) Debug(msg string, args ...interface{}) {
	z.sugar.Debugw(msg, args...)
}

func (z *ZapLogger) Info(msg string, args ...interface{}) {
	z.sugar.Infow(msg, args...)
}

func (z *ZapLogger) Warn(msg string, args ...interface{}) {
	z.sugar.Warnw(msg, args...)
}

func (z *ZapLogger) Error(msg string, args ...interface{}) {
	z.sugar.Errorw(msg, args...)
}

func (z *ZapLogger) With(args ...interface{}) Logger {
	return &ZapLogger{
		logger: z.logger,
		sugar:  z.sugar.With(args...),
	}
}

func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	return &ZapLogger{
		logger: z.logger.With(zap.Any("context", ctx)),
		sugar:  z.sugar.With("context", ctx),
	}
}
