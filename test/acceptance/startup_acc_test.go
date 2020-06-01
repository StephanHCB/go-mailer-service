package acceptance

import (
	"github.com/StephanHCB/go-mailer-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestStartupWithValidConfig_ShouldBeHealthy(t *testing.T) {
	docs.Given("Given a valid configuration")
	configAndSecretsPath := tstValidConfigurationPath

	docs.When("When the application is started")
	tstSetup(configAndSecretsPath)
	defer tstShutdown()

	docs.Then("Then it reports as healthy")
	response, err := tstPerformGet("/health", tstUnauthenticated())

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, response.status)
}
