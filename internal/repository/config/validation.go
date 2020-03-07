package config

import (
	"fmt"
	"github.com/spf13/viper"
)

func checkLength(min int, max int, key string) error {
	value := viper.GetString(key)
	if len(value) < min || len(value) > max {
		return fmt.Errorf("Fatal error: configuration value for key %s must be between %d and %d characters long\n", key, min, max)
	}
	return nil
}

func checkValidPortNumber(key string) error {
	port := viper.GetUint(key)
	if port < 1024 || port > 65535 {
		return fmt.Errorf("Fatal error: configuration value for key %s is not in range 1024..65535\n", key)
	}
	return nil
}
