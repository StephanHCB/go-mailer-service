package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
)


func Setup() {
	// configure to implement a small subset of ECS as an example
	// see https://www.elastic.co/guide/en/ecs/1.4

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.LevelFieldName = "log.level"
	zerolog.MessageFieldName = "message" // correct by default

	// assume JSON logging at first, until config is loaded
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func PostConfigSetup() {
	profiles := viper.GetStringSlice("profiles")
	if contains(profiles, "local") {
		// switch to developer friendly colored console log
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		// stay with JSON logging
		serviceName := viper.GetString("service.name")
		log.Logger = log.With().Str("service.id", serviceName).Logger()
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}