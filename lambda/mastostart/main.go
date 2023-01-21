package main

import (
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rmrfslashbin/mastostart/pkg/app"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rs/zerolog"
)

func main() {
	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Fetch the log level from the environment
	logLevel := os.Getenv("LOGLEVEL")

	// Set the log level
	switch strings.ToLower(logLevel) {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Msg("startup!")

	db, err := database.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database")
	}

	if a, err := app.New(
		app.WithDB(db),
		app.WithLogger(&log),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to create app")
	} else {
		lambda.Start(a.LambdaHandler)
	}
}
