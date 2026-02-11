package log

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(lavel string) *Logger {
	lvl, err := zerolog.ParseLevel(lavel)

	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.GlobalLevel()

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	if lvl <= zerolog.DebugLevel {
		output.NoColor = false
		output.FormatLevel = func(i interface{}) string {
			return i.(string)
		}
	}

	l := zerolog.New(output).
		Level(lvl).
		With().
		Timestamp().
		Str("service", "hookord").
		Str("version", "0.0.1").
		Logger()

	return &Logger{
		logger: &l,
	}
}

func (l *Logger) GetLogger() *zerolog.Logger {
	return l.logger
}

func (l *Logger) WithRequestId(reqId string) zerolog.Logger {
	return l.logger.With().Str("request_id", reqId).Logger()
}

func (l *Logger) Caller() *zerolog.Event {
	_, file, line, _ := runtime.Caller(1)
	logger := l.logger.
		With().
		Str("caller", file+":"+strconv.Itoa(line)).
		Logger()
	return logger.
		Info()
}
