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

	for _, item := range configItems {
		// must set null default values here, or else this value will overwrite config values from config file
		if _, ok := item.Default.(string); ok {
			pflag.String(item.Key, "", item.Description)
		}
		if _, ok := item.Default.(uint); ok {
			pflag.Uint(item.Key, 0, item.Description)
		}
		if _, ok := item.Default.([]string); ok {
			pflag.StringSlice(item.Key, []string{}, item.Description)
		}
	}
}

func setupFlags() {
	for _, item := range configItems {
		bindFlag(item.Key)
	}
}

func bindFlag(key string) {
	err := viper.BindPFlag(key, pflag.Lookup(key))
	if err != nil {
		failFunction(fmt.Errorf("Fatal error could not bind configuration flag %s: %s\n", key, err))
	}
}
