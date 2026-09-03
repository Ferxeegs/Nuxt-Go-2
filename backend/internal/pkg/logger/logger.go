package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Init configures the global zerolog logger with the given level and mode.
// In production it emits structured JSON; in development it uses a
// human-readable console writer.
func Init(level string, production bool) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	var l zerolog.Logger
	if production {
		l = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		l = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	}

	zerolog.DefaultContextLogger = &l
	return l
}
