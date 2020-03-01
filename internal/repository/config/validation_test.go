package config

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

var recordedErrors []error

func tstSetup(address string, port uint) {
	viper.SetDefault(configKeyServerAddress, address)
	viper.SetDefault(configKeyServerPort, port)
	recordedErrors = nil
	failFunction = func(err error) {
		recordedErrors = append(recordedErrors, err)
	}
}

func TestCheckLength_EmptyOkIfMinZero(t *testing.T) {
	tstSetup("", 8080)

	checkLength(0, 20, configKeyServerAddress)
	require.Nil(t, recordedErrors)
}

func TestCheckLength_ShouldAllowAcceptableLength(t *testing.T) {
	tstSetup("1234567890", 8080)

	checkLength(5, 20, configKeyServerAddress)
	require.Nil(t, recordedErrors)
}

func TestCheckLength_ShouldFailIfTooLong(t *testing.T) {
	tstSetup("123456789012345678901", 8080)

	checkLength(0, 20, configKeyServerAddress)
	expectedMessage := "Fatal error: configuration value for key server.address must be between 0 and 20 characters long\n"
	require.Equal(t, 1, len(recordedErrors))
	require.Equal(t, expectedMessage, recordedErrors[0].Error())
}

func TestCheckValidPortNumber_Ok(t *testing.T) {
	tstSetup("", 8080)

	checkValidPortNumber(configKeyServerPort)
	require.Nil(t, recordedErrors)
}

func TestCheckValidPortNumber_TooLow(t *testing.T) {
	tstSetup("", 443)

	checkValidPortNumber(configKeyServerPort)
	expectedMessage := "Fatal error: configuration value for key server.port is not in range 1024..65535\n"
	require.Equal(t, 1, len(recordedErrors))
	require.Equal(t, expectedMessage, recordedErrors[0].Error())
}

func TestCheckValidPortNumber_TooHigh(t *testing.T) {
	tstSetup("", 65536)

	checkValidPortNumber(configKeyServerPort)
	expectedMessage := "Fatal error: configuration value for key server.port is not in range 1024..65535\n"
	require.Equal(t, 1, len(recordedErrors))
	require.Equal(t, expectedMessage, recordedErrors[0].Error())
}
