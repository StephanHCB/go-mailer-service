package main

import (
	_ "github.com/StephanHCB/go-mailer-service/docs"
	"github.com/StephanHCB/go-mailer-service/internal/repository/config"
	"github.com/StephanHCB/go-mailer-service/web"
	"github.com/spf13/pflag"
)

func main() {
	pflag.Parse()
	config.Setup()
	web.Serve()
}
