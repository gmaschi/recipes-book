package random

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestString(t *testing.T) {
	n := 10
	str := String(n)

	require.IsType(t, "", str)
	require.Len(t, str, n)
}

func TestEmail(t *testing.T) {
	email := Email()

	require.IsType(t, "", email)
	require.Contains(t, email, "@")
}

func TestStringSlice(t *testing.T) {
	n := 5
	ss := StringSlice(n)

	require.IsType(t, []string{}, ss)
	require.Len(t, ss, n)
}