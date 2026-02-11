package log

import (
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(lavel string) *Logger {
	lvl, err := zerolog.ParseLevel(lavel)

	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeformatUnix
}
