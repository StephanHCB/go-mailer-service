package config

import "github.com/spf13/viper"

const configEnvServerAddress = "CONFIG_SERVER_ADDRESS"
const configEnvServerPort = "CONFIG_SERVER_PORT"

func setupEnv() {
	// the only error that can occur is when the Key is empty
	_ = viper.BindEnv(configKeyServerAddress, configEnvServerAddress)
	_ = viper.BindEnv(configKeyServerPort, configEnvServerPort)
}
