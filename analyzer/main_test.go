package analyzer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain_UsageError(t *testing.T) {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer"}, &out, &err)
	require.Equal(t, 2, code)
	require.Contains(t, err.String(), "usage:")
}

func TestMain_NonexistentChart(t *testing.T) {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer", "does-not-exist"}, &out, &err)
	require.Equal(t, 1, code)
	require.Contains(t, err.String(), "error:")
}

func TestMain_SimpleChart(t *testing.T) {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer", "../test-chart"}, &out, &err)
	require.Equal(t, 0, code, err.String())

	s := out.String()
	require.Contains(t, s, "Referenced:")
	require.True(t, strings.Contains(s, "config.enabled") && strings.Contains(s, "config.message"))
	require.Contains(t, s, "Defined-not-used:")
	require.Contains(t, s, "Used-not-defined:")
	require.Equal(t, "", err.String())
}
