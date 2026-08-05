package logger

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestSetup_LogLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel zerolog.Level
	}{
		{
			name:          "valid debug level",
			level:         "debug",
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "valid warn level",
			level:         "warn",
			expectedLevel: zerolog.WarnLevel,
		},
		{
			name:          "valid error level",
			level:         "error",
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid_level",
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "empty level defaults to NoLevel",
			level:         "",
			expectedLevel: zerolog.NoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(Config{
				Level:       tt.level,
				Environment: "development",
			})

			if zerolog.GlobalLevel() != tt.expectedLevel {
				t.Errorf("GlobalLevel() = %v, want %v", zerolog.GlobalLevel(), tt.expectedLevel)
			}

			if zerolog.TimeFieldFormat != time.RFC3339 {
				t.Errorf("TimeFieldFormat = %v, want %v", zerolog.TimeFieldFormat, time.RFC3339)
			}
		})
	}
}

func TestSetup_Environments(t *testing.T) {
	tests := []struct {
		name        string
		environment string
	}{
		{
			name:        "production environment",
			environment: "production",
		},
		{
			name:        "kubernetes environment",
			environment: "kubernetes",
		},
		{
			name:        "development environment",
			environment: "development",
		},
		{
			name:        "other environment defaults to console writer",
			environment: "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(Config{
				Level:       "info",
				Environment: tt.environment,
			})

			log.Logger.Info().Msg("testing environment log setup")
		})
	}
}
