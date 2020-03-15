package configuration

import (
	"github.com/StephanHCB/go-autumn-config"
	"github.com/rs/zerolog/log"
)

// initialize configuration with full setup - you need to call this
func Setup() {
	auconfig.Setup(configItems, fail, warn)
	auconfig.Load()
}

func fail(err error) {
	// this will os.exit 1
	log.Fatal().Err(err).Msg(err.Error())
}

func warn(message string) {
	log.Warn().Msg(message)
}
