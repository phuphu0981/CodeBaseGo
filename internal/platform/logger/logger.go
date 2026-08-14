package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"codebasego/internal/platform/config"
)

func NewLogger(cfg *config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.DebugLevel
	}

	return zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
	).Level(level).With().Timestamp().Caller().Logger()
}
