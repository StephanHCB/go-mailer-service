package configuration

import (
	"github.com/StephanHCB/go-autumn-config"
	"github.com/StephanHCB/go-autumn-config-api"
)

const configKeyServerAddress = "server.address"
const configKeyServerPort = "server.port"
const configKeyServiceName = "service.name"

var configItems = []auconfigapi.ConfigItem{
	auconfig.ConfigItemProfile,
	{
		Key:         configKeyServerAddress,
		Default:     "",
		Description: "ip address or hostname to listen on, can be left blank for localhost",
		Validate:    func(key string) error { return checkLength(0, 255, key) },
	}, {
		Key:         configKeyServerPort,
		Default:     uint(8080),
		Description: "port to listen on, defaults to 8080 if not set",
		Validate:    checkValidPortNumber,
	}, {
		Key:         configKeyServiceName,
		Default:     "unnamed-service",
		Description: "name of service, used for logging",
		Validate:    func(key string) error { return checkLength(1, 255, key) },
	},
}
