package main

import (
	_ "github.com/StephanHCB/go-mailer-service/docs"
	"github.com/StephanHCB/go-mailer-service/internal/repository/config"
	"github.com/StephanHCB/go-mailer-service/internal/repository/logging"
	"github.com/StephanHCB/go-mailer-service/web"
)

func main() {
	logging.Setup()
	config.Setup()
	logging.PostConfigSetup()

	web.Serve()
}
