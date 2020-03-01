package config

import "github.com/spf13/viper"

const configKeyServerAddress = "server.address"
const configDefaultServerAddress = ""
const configDescriptionServerAddress = "ip address or hostname to listen on, can be left blank for localhost"

const configKeyServerPort = "server.port"
const configDefaultServerPort uint = 8080
const ConfigDescriptionServerPort = "port to listen on, defaults to 8080 if not set"

func setupDefaults() {
	viper.SetDefault(configKeyServerAddress, configDefaultServerAddress)
	viper.SetDefault(configKeyServerPort, configDefaultServerPort)
}
