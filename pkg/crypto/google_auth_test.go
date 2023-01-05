package crypto_test

import (
	"testing"

	"github.com/limingyao/excellent-go/pkg/crypto"
)

func TestGoogleAuth_GetSecret(t *testing.T) {
	g := crypto.NewGoogleAuth()
	t.Log(g.GetSecret())
}
