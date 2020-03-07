package web

import (
	"fmt"
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/StephanHCB/go-mailer-service/web/controller/emailctl"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// use this for easy mocking

var failFunction = fail

func Serve() {
	// turn off annoying printf logging from gin
	gin.SetMode(gin.ReleaseMode)

	server := gin.New()
	server.Use(logger.SetLogger(), gin.Recovery())

	_ = emailctl.Create(server)

	// TODO move this out to a static files controller
	// serve swagger-ui and swagger.json
	server.Static("/swagger-ui/", "third_party/swagger_ui")
	server.StaticFile("swagger.json", "docs/swagger.json")

	address := configuration.ServerAddress()
	log.Info().Msg("Starting web server on " + address)
	err := server.Run(address)
	if err != nil {
		// TODO log a warning on intentional shutdown, and an error otherwise
		failFunction(fmt.Errorf("Fatal error while starting web server: %s\n", err))
	}
}

func fail(err error) {
	log.Fatal().Err(err)
}
