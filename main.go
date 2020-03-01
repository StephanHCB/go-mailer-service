package main

import (
	"github.com/StephanHCB/go-mailer-service/internal/repository/config"
	"github.com/spf13/pflag"
)

func main() {
	pflag.Parse()
	config.Setup()
}
