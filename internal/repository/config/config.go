package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// public functions for accessing all configuration values.

func ServerAddress() string {
	return fmt.Sprintf("%v:%d", viper.GetString(configKeyServerAddress), viper.GetUint(configKeyServerPort))
}

// initialize configuration with full setup - you need to call this
func Setup() {
	setupDefaults()
	setupLoading()
	setupEnv()
	setupFlags()
	validate()
}
