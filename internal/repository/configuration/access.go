package configuration

import (
	"fmt"
	"github.com/spf13/viper"
)

// public functions for accessing all configuration values.

func ServerAddress() string {
	return fmt.Sprintf("%v:%d", viper.GetString(configKeyServerAddress), viper.GetUint(configKeyServerPort))
}

func ServiceName() string {
	return viper.GetString(configKeyServiceName)
}

func IsProfileActive(profileName string) bool {
	profiles := viper.GetStringSlice("profiles")
	return contains(profiles, profileName)
}

func SecuritySecret() string {
	return viper.GetString(configKeySecuritySecret)
}
