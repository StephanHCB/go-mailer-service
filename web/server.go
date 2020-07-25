package web

import (
	"fmt"
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/StephanHCB/go-mailer-service/internal/service/emailsrv"
	"github.com/StephanHCB/go-mailer-service/web/controller/emailctl"
	"github.com/StephanHCB/go-mailer-service/web/controller/healthctl"
	"github.com/StephanHCB/go-mailer-service/web/controller/swaggerctl"
	"github.com/StephanHCB/go-mailer-service/web/middleware/authentication"
	"github.com/StephanHCB/go-mailer-service/web/middleware/ctxlogger"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	requestid "github.com/thanhhh/gin-requestid"
)

// use this for easy mocking

var failFunction = fail

func Create() *gin.Engine {
	// turn off annoying printf logging from gin
	gin.SetMode(gin.ReleaseMode)

	server := gin.New()
	server.Use(requestid.RequestID(),
		logger.SetLogger(),
		ctxlogger.AddZerologLoggerToRequestContext(),
		// TODO secret should come from configuration
		authentication.AddJWTTokenInfoToContextHandlerFunc(configuration.SecuritySecret()),
		gin.Recovery())

	return server
}

func AddRoutes(server *gin.Engine, emailService emailsrv.EmailService) {
	_ = emailctl.Create(server, emailService)

	healthctl.Create(server)

	swaggerctl.SetupSwaggerRoutes(server)
}

func Serve() {
	server := Create()

	AddRoutes(server, emailsrv.Create())

	address := configuration.ServerAddress()
	log.Info().Msg("Starting web server on " + address)
	err := server.Run(address)
	if err != nil {
		// TODO log a warning on intentional shutdown, and an error otherwise
		failFunction(fmt.Errorf("Fatal error while starting web server: %s\n", err))
	}
}

func fail(err error) {
	log.Fatal().Err(err).Msg(err.Error())
}
