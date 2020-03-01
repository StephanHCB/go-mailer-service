package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// use this for easy mocking

var failFunction = fail
var warnFunction = warn

func validate() {
	checkLength(0, 255, configKeyServerAddress)
	checkValidPortNumber(configKeyServerPort)
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
	// TODO fatal logging and proper application stop
	panic(err)
}

func warn(message string) {
	// TODO log a warning with this message
}