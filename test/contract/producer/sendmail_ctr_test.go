package producer

import (
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

// see setup_ctr_test.go for http test server and service mock

// tests are run in the directory they are located in
// normally we would use a web server to which we publish the contracts, but this is fine for this example
const consumerPactDir = "../../../../go-campaign-service/test/contract/consumer/pacts/"

// contract test provider side

func TestProvider(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "CampaignService",
		Provider: "MailService",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// provider API is already running
	// (done during test startup using httptest package, see setup_ctr_test.go)

	// Verify the Provider using the locally saved Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: ts.URL,
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/campaignservice-mailservice.json", consumerPactDir))},
		StateHandlers: 			types.StateHandlers{
			// Setup any state required by the test
			// example that we are not really using in this test
			"an authorized user with the admin role exists": func() error {
				// e.g. set up service mock responses here if needed
				return nil
			},
		},
	})
	require.Nil(t, err, "unexpected error during verification")
	// now use the service mock to assert further expectations of what calls to the mock service should have
	// occurred during the verification.
}
