package password

import (
	"github.com/gmaschi/go-recipes-book/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPassword(t *testing.T) {
	t.Run("Correct password case", func(t *testing.T) {
		password := random.String(8)

		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)

		err = CheckPassword(password, hashedPassword)
		require.NoError(t, err)
	})

	t.Run("Wrong password case", func(t *testing.T) {
		password := random.String(8)
		wrongPassword := random.String(10)

		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)

		err = CheckPassword(wrongPassword, hashedPassword)
		require.Error(t, err)
		require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})

	t.Run("Different hashes for the same password", func(t *testing.T) {
		password := random.String(8)

		hashedPassword1, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword1)

		hashedPassword2, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword2)

		err = CheckPassword(password, hashedPassword1)
		require.NoError(t, err)

		err = CheckPassword(password, hashedPassword2)
		require.NoError(t, err)

		require.NotEqual(t, hashedPassword1, hashedPassword2)
	})
}
