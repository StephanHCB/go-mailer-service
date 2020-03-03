package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type configItem struct {
	Key         string
	Default     interface{}
	Description string
	EnvName     string
	Validate    func(key string)
}

// initialize configuration with full setup - you need to call this
func Setup() {
	pflag.Parse()

	setupDefaults()
	setupLoading()
	setupEnv()
	setupFlags()
	validate()
}

func setupDefaults() {
	for _, item := range configItems {
		viper.SetDefault(item.Key, item.Default)
	}
}

func setupEnv() {
	for _, item := range configItems {
		// the only error that can occur is when the Key is empty
		_ = viper.BindEnv(item.Key, item.EnvName)
	}
}

