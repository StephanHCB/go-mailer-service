package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configPath string
var secretsPath string

func init() {
	pflag.StringVar(&configPath, "config-path", "", "config file path without file name")
	pflag.StringVar(&secretsPath, "secrets-path", "", "secrets file path without file name")
	pflag.String(configKeyServerAddress, "", configDescriptionServerAddress)
	pflag.Uint(configKeyServerPort, 0, ConfigDescriptionServerPort)
}

func setupFlags() {
	bindFlag(configKeyServerAddress)
	bindFlag(configKeyServerPort)
}

func bindFlag(key string) {
	err := viper.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		failFunction(fmt.Errorf("Fatal error could not bind configuration flag %s: %s\n", key, err))
	}
}
