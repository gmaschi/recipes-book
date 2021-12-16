package validators

import (
	"github.com/gmaschi/go-recipes-book/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPassword(t *testing.T) {
	t.Run("Valid password", func(t *testing.T) {
		password := random.String(10)
		require.True(t, Password(password))
	})

	t.Run("Invalid password", func(t *testing.T) {
		password := random.String(4)
		require.False(t, Password(password))
	})
}

func TestEmail(t *testing.T) {
	t.Run("Valid email", func(t *testing.T) {
		email := random.Email()
		require.True(t, Email(email))
	})

	t.Run("Invalid email", func(t *testing.T) {
		email := random.String(16)
		require.False(t, Email(email))
	})
}
