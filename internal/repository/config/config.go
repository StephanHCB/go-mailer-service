package config

const configKeyServerAddress = "server.address"
const configKeyServerPort = "server.port"
const configKeyServiceName = "service.name"
const configKeyProfiles = "profiles"

var configItems = []configItem{
	{
		Key:         configKeyServerAddress,
		Default:     "",
		Description: "ip address or hostname to listen on, can be left blank for localhost",
		EnvName:     "CONFIG_SERVER_ADDRESS",
		Validate:    func(key string) { checkLength(0, 255, key) },
	}, {
		Key:         configKeyServerPort,
		Default:     uint(8080),
		Description: "port to listen on, defaults to 8080 if not set",
		EnvName:     "CONFIG_SERVER_PORT",
		Validate:    checkValidPortNumber,
	}, {
		Key:         configKeyServiceName,
		Default:     "unnamed-service",
		Description: "name of service, used for logging",
		EnvName:     "CONFIG_SERVICE_NAME",
		Validate:    func(key string) { checkLength(1, 255, key) },
	}, {
		Key:         configKeyProfiles,
		Default:     []string{},
		Description: "list of profiles, separate by spaces in environment or command line parameters",
		EnvName:     "CONFIG_PROFILES",
		Validate:    func(_ string) {},
	},
}
