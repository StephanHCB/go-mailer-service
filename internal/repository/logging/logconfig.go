package logging

import (
	"bytes"
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)


func Setup() {
	// configure to implement a small subset of ECS as an example
	// see https://www.elastic.co/guide/en/ecs/1.4

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.LevelFieldName = "log.level"
	zerolog.MessageFieldName = "message" // correct by default

	// assume JSON logging at first, until configuration is loaded
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func PostConfigSetup() {
	if configuration.IsProfileActive("local") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		log.Info().Msg("switching to developer friendly console log because profile 'local' is active")
	} else {
		// stay with JSON logging and add ECS service.id field
		log.Logger = log.With().Str("service.id", configuration.ServiceName()).Logger()
	}
}

var RecordedLogForTesting = new(bytes.Buffer)

// alternative Setup function for testing that records log entries instead of writing them to console
func SetupForTesting() {
	Setup()
	log.Logger = zerolog.New(RecordedLogForTesting).With().Timestamp().Logger()
}
