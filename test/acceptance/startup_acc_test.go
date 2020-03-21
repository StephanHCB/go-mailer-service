package acceptance

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

func TestStartupWithValidConfig_ShouldBeHealthy(t *testing.T) {
	Convey("Given a valid configuration", t, func() {
		configAndSecretsPath := tstValidConfigurationPath

		Convey("When the application is started", func() {
			tstSetup(configAndSecretsPath)
			defer tstShutdown()

			Convey("Then it reports as healthy", func() {
				response, err := tstPerformGet("/health", tstUnauthenticated())

				So(err, ShouldEqual, nil)
				So(response.status, ShouldEqual, http.StatusOK)
			})
		})
	})
}
