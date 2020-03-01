package config

import (
	"fmt"
	"github.com/spf13/viper"
)

const configFileName = "config"
const secretsFileName = "secrets"

func setupLoading() {
	if configPath == "" {
		warnFunction("you did not provide the config-path command line flag. Falling back to looking for config.(yaml|json) in current directory.")
		configPath = "."
	}
	viper.SetConfigName(configFileName)
	viper.AddConfigPath(configPath)
	handleLoadingError(viper.ReadInConfig(), configFileName, configPath, false)

	if secretsPath != "" {
		viper.SetConfigName(secretsFileName)
		viper.AddConfigPath(secretsPath)
		handleLoadingError(viper.MergeInConfig(), secretsFileName, secretsPath, true)
	} else {
		warnFunction("you did not provide the secrets-path command line flag. No secrets file will be loaded. This may be ok on a local machine.")
	}
}

func handleLoadingError(err error, name string, path string, suppressErrorMessage bool) {
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			failFunction(fmt.Errorf("Fatal error: configuration file %s.(yaml|json) not found in %s: %s\n", name, path, err))
		} else {
			if suppressErrorMessage {
				// do not print the actual error as it may contain snippets from the secrets file
				failFunction(fmt.Errorf("Fatal error: configuration file %s.(yaml|json) found but failed to load. Hiding error message because this is a secrets file\n", name))
			} else {
				failFunction(fmt.Errorf("Fatal error: configuration file %s.(yaml|json) found but failed to load: %s\n", name, err))
			}
		}
	}
}
