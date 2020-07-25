package acceptance

import (
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/StephanHCB/go-mailer-service/internal/repository/logging"
	"github.com/StephanHCB/go-mailer-service/internal/service/emailsrv"
	"github.com/StephanHCB/go-mailer-service/web"
	"net/http/httptest"
)

// placing these here because they are package global

var (
	ts *httptest.Server
	failures []error
	warnings []string
)

const tstValidConfigurationPath =  "../resources/validconfig"

func tstSetup(configAndSecretsPath string) {
	tstSetupConfig(configAndSecretsPath, configAndSecretsPath)
	if !tstHadFailures() {
		tstSetupHttpTestServer()
	}
}

func tstFail(err error) {
	failures = append(failures, err)
}

func tstWarn(msg string) {
	warnings = append(warnings, msg)
}

func tstSetupConfig(configPath string, secretsPath string) {
	failures = []error{}
	warnings = []string{}
	logging.SetupForTesting()
	configuration.SetupForIntegrationTest(tstFail, tstWarn, configPath, secretsPath)
	if !tstHadFailures() {
		logging.PostConfigSetup()
	}
}

func tstHadFailures() bool {
	return len(failures) > 0
}

func tstSetupHttpTestServer() {
	router := web.Create()
	web.AddRoutes(router, emailsrv.Create())
	ts = httptest.NewServer(router)
}

func tstShutdown() {
	if !tstHadFailures() {
		ts.Close()
	}
}
