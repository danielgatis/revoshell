package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger zerolog.Logger
)

// Init initializes the logger with pretty console output.
func Init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	Logger = zerolog.New(output).With().Timestamp().Logger()
	log.Logger = Logger
}

// InitWithComponent initializes logger with a component name.
func InitWithComponent(component string) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}

	return zerolog.New(output).
		With().
		Timestamp().
		Str("component", component).
		Logger()
}
