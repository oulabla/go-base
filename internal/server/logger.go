package server

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(debug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true

	var output io.Writer = os.Stdout // базовый writer

	if debug {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05.000",
			NoColor:    false, // можно true в CI/CD
		}
	}

	log.Logger = zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()

	if debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	zerolog.DefaultContextLogger = &log.Logger
}
