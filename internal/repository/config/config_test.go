package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServerAddress(t *testing.T) {
	setupDefaults()

	expected := ":8080"
	actual := ServerAddress()
	require.Equal(t, expected, actual)
}
