package acceptance

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TODO implement these once we have a web framework

func TestStartupWithValidConfig_ShouldBeHealthy(t *testing.T) {
	Convey("Given a valid configuration", t, func() {
		x := 1

		Convey("When the application is started", func() {
			x++

			Convey("Then it reports as healthy", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}

func TestStartupWithInvalidConfig_ShouldNotRespond(t *testing.T) {
	Convey("Given an invalid configuration", t, func() {
		x := 1

		Convey("When the application is started", func() {
			x++

			Convey("Then it does not respond to health inquiries", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}
