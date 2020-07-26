package main

import (
	_ "github.com/StephanHCB/go-mailer-service/docs"
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/StephanHCB/go-mailer-service/internal/repository/logging"
	"github.com/StephanHCB/go-mailer-service/internal/repository/metricspush"
	"github.com/StephanHCB/go-mailer-service/web"
)

func main() {
	logging.Setup()
	configuration.Setup()
	logging.PostConfigSetup()
	metricspush.Setup()

	web.Serve()
}
