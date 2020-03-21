package configuration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContains_yes_first(t *testing.T) {
	actual := contains([]string{"development", "squirrel", "local"}, "development")
	require.True(t, actual)
}

func TestContains_yes_last(t *testing.T) {
	actual := contains([]string{"development", "squirrel", "local"}, "local")
	require.True(t, actual)
}

func TestContains_yes_only(t *testing.T) {
	actual := contains([]string{"local"}, "local")
	require.True(t, actual)
}

func TestContains_no_empty(t *testing.T) {
	actual := contains([]string{}, "local")
	require.False(t, actual)
}

func TestContains_no(t *testing.T) {
	actual := contains([]string{"development", "local"}, "cat")
	require.False(t, actual)
}

