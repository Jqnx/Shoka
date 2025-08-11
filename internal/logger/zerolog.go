package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

//var once sync.Once
//
//var log zerolog.Logger
//
//func NewZeroLog() *zerolog.Logger {
//	once.Do(func() {
//		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
//		zerolog.TimeFieldFormat = time.RFC3339Nano
//
//		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
//		if err != nil {
//			logLevel = int(zerolog.InfoLevel) // default to INFO
//		}
//
//		var output io.Writer = zerolog.ConsoleWriter{
//			Out:        os.Stdout,
//			TimeFormat: time.RFC3339,
//		}
//
//		// if os.Getenv("APP_ENV") != "development" {
//		//	fileLogger := &lumberjack.Logger{
//		//		Filename:   "shoka.log",
//		//		MaxSize:    5, //
//		//		MaxBackups: 10,
//		//		MaxAge:     14,
//		//		Compress:   true,
//		//	}
//
//		//	output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
//		//}
//
//		log = zerolog.New(output).
//			Level(zerolog.Level(logLevel)).
//			With().
//			Timestamp().
//			Caller().
//			Logger()
//	})
//
//	return &log
//}

// ZerologLogger wraps zerolog.Logger
type ZerologLogger struct {
	logger zerolog.Logger
}

// NewZerologLogger creates a new ZerologLogger instance
func NewZerologLogger(logger zerolog.Logger) *ZerologLogger {
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	if logger.GetLevel() == 0 {
		logger = zerolog.New(output).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	}
	return &ZerologLogger{logger: logger}
}

func (zl *ZerologLogger) Debug(msg string, args ...interface{}) {
	event := zl.logger.Debug()
	zl.addFields(event, args...)
	event.Msg(msg)
}

func (zl *ZerologLogger) Info(msg string, args ...interface{}) {
	event := zl.logger.Info()
	zl.addFields(event, args...)
	event.Msg(msg)
}

func (zl *ZerologLogger) Warn(msg string, args ...interface{}) {
	event := zl.logger.Warn()
	zl.addFields(event, args...)
	event.Msg(msg)
}

func (zl *ZerologLogger) Error(msg string, args ...interface{}) {
	event := zl.logger.Error()
	zl.addFields(event, args...)
	event.Msg(msg)
}

func (zl *ZerologLogger) Fatal(msg string, args ...interface{}) {
	event := zl.logger.Fatal()
	zl.addFields(event, args...)
	event.Msg(msg)
}

func (zl *ZerologLogger) With(args ...interface{}) Logger {
	ctx := zl.logger.With()
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := args[i].(string)
			value := args[i+1]
			ctx = ctx.Interface(key, value)
		}
	}
	return &ZerologLogger{logger: ctx.Logger()}
}

func (zl *ZerologLogger) WithContext(ctx context.Context) Logger {
	return &ZerologLogger{
		logger: zl.logger.With().Interface("context", ctx).Logger(),
	}
}

func (zl *ZerologLogger) addFields(event *zerolog.Event, args ...interface{}) {
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := args[i].(string)
			value := args[i+1]
			event.Interface(key, value)
		}
	}
}
