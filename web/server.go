package web

import (
	"fmt"
	"github.com/StephanHCB/go-mailer-service/internal/repository/config"
	"github.com/StephanHCB/go-mailer-service/web/controller/emailctl"
	"github.com/gin-gonic/gin"
)

// use this for easy mocking

var failFunction = fail

func Serve() {
	server := gin.Default()

	_ = emailctl.Create(server)

	err := server.Run(config.ServerAddress())
	if err != nil {
		// TODO log a warning on intentional shutdown, and an error otherwise
		failFunction(fmt.Errorf("Fatal error while starting web server: %s\n", err))
	}
}

func fail(err error) {
	// TODO fatal logging and proper application stop
	panic(err)
}
