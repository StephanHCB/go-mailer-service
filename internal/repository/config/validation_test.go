package config

import (
	"github.com/StephanHCB/go-autumn-config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

var recordedErrors []error
var recordedWarnings []string

var failFunction = func(err error) {
	recordedErrors = append(recordedErrors, err)
}
var warnFunction = func(msg string) {
	recordedWarnings = append(recordedWarnings, msg)
}

func tstSetup(address string, port uint) {
	recordedErrors = nil
	recordedWarnings = nil
	auconfig.SetupDefaultsOnly(configItems, failFunction, warnFunction)
	viper.Set(configKeyServerAddress, address)
	viper.Set(configKeyServerPort, port)
}

func TestCheckLength_EmptyOkIfMinZero(t *testing.T) {
	tstSetup("", 8080)

	err := checkLength(0, 20, configKeyServerAddress)
	require.Nil(t, err)
}

func TestCheckLength_ShouldAllowAcceptableLength(t *testing.T) {
	tstSetup("1234567890", 8080)

	err := checkLength(5, 20, configKeyServerAddress)
	require.Nil(t, err)
}

func TestCheckLength_ShouldFailIfTooLong(t *testing.T) {
	tstSetup("123456789012345678901", 8080)

	err := checkLength(0, 20, configKeyServerAddress)
	expectedMessage := "Fatal error: configuration value for key server.address must be between 0 and 20 characters long\n"
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())
}

func TestCheckValidPortNumber_Ok(t *testing.T) {
	tstSetup("", 8080)

	err := checkValidPortNumber(configKeyServerPort)
	require.Nil(t, err)
}

func TestCheckValidPortNumber_TooLow(t *testing.T) {
	tstSetup("", 443)

	err := checkValidPortNumber(configKeyServerPort)
	expectedMessage := "Fatal error: configuration value for key server.port is not in range 1024..65535\n"
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())
}

func TestCheckValidPortNumber_TooHigh(t *testing.T) {
	tstSetup("", 65536)

	err := checkValidPortNumber(configKeyServerPort)
	expectedMessage := "Fatal error: configuration value for key server.port is not in range 1024..65535\n"
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())
}
