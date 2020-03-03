package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// use this for easy mocking

var failFunction = fail
var warnFunction = warn

func validate() {
	for _, item := range configItems {
		item.Validate(item.Key)
	}
}

func checkLength(min int, max int, key string) {
	value := viper.GetString(key)
	if len(value) < min || len(value) > max {
		failFunction(fmt.Errorf("Fatal error: configuration value for key %s must be between %d and %d characters long\n", key, min, max))
	}
}

func checkValidPortNumber(key string) {
	port := viper.GetUint(key)
	if port < 1024 || port > 65535 {
		failFunction(fmt.Errorf("Fatal error: configuration value for key %s is not in range 1024..65535\n", key))
	}
}

func fail(err error) {
	// this will os.exit 1
	log.Fatal().Err(err)
}

func warn(message string) {
	log.Warn().Msg(message)
}
