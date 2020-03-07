package configuration

import (
	"github.com/StephanHCB/go-autumn-config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServerAddress(t *testing.T) {
	auconfig.SetupDefaultsOnly(configItems, failFunction, warnFunction)

	expected := ":8080"
	actual := ServerAddress()
	require.Equal(t, expected, actual)
}
