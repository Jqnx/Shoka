package logger

import (
	"context"

	"github.com/rs/zerolog"
)

// Logger defines a common interface for structured logging
// Uses variadic interface{} args in key-value pairs
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	With(args ...interface{}) Logger
	WithContext(ctx context.Context) Logger
}

// Helper function to convert variadic args to key-value pairs
// Panics if odd number of arguments
func argsToKeyValues(args []interface{}) ([]interface{}, error) {
	if len(args)%2 != 0 {
		// Could return error instead of panic for production use
		panic("odd number of arguments - must provide key-value pairs")
	}
	return args, nil
}

// NewLogger creates a logger instance based on the specified type
func NewLogger(loggerType string) Logger {
	switch loggerType {
	case "slog":
		return NewSlogLogger(nil)
	case "zap":
		return NewZapLogger(nil)
	case "zerolog":
		return NewZerologLogger(zerolog.Logger{})
	default:
		return NewSlogLogger(nil) // Default to slog
	}
}

// TODO: Look at config based logger levels
// Configuration-based logger creation
//type Config struct {
//	Type  LoggerType `json:"type" yaml:"type"`
//	Level string     `json:"level" yaml:"level"`
//}
//
//func NewLoggerFromConfig(config Config) Logger {
//	switch config.Type {
//	case SlogType:
//		var level slog.Level
//		switch config.Level {
//		case "debug":
//			level = slog.LevelDebug
//		case "info":
//			level = slog.LevelInfo
//		case "warn":
//			level = slog.LevelWarn
//		case "error":
//			level = slog.LevelError
//		default:
//			level = slog.LevelInfo
//		}
//		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
//		return NewSlogLogger(logger)
//	case ZapType:
//		config := zap.NewProductionConfig()
//		switch config.Level {
//		case "debug":
//			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
//		case "info":
//			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
//		case "warn":
//			config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
//		case "error":
//			config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
//		}
//		logger, _ := config.Build()
//		return NewZapLogger(logger)
//	case ZerologType:
//		var level zerolog.Level
//		switch config.Level {
//		case "debug":
//			level = zerolog.DebugLevel
//		case "info":
//			level = zerolog.InfoLevel
//		case "warn":
//			level = zerolog.WarnLevel
//		case "error":
//			level = zerolog.ErrorLevel
//		default:
//			level = zerolog.InfoLevel
//		}
//		logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
//		return NewZerologLogger(logger)
//	default:
//		return NewLogger(SlogType)
//	}
//}
