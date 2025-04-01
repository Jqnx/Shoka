package logger

import (
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Logger struct {
	Logger *zerolog.Logger
}

var once sync.Once

var log zerolog.Logger

func New() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		// if os.Getenv("APP_ENV") != "development" {
		//	fileLogger := &lumberjack.Logger{
		//		Filename:   "shoka.log",
		//		MaxSize:    5, //
		//		MaxBackups: 10,
		//		MaxAge:     14,
		//		Compress:   true,
		//	}

		//	output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		//}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return log
}

func Get() Logger {
	logger := New()
	return Logger{
		Logger: &logger,
	}
}
