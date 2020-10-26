package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	token := SecureToken()
	require.NotNil(t, token)
}
