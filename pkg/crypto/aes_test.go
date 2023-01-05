package crypto_test

import (
	"testing"

	"github.com/limingyao/excellent-go/pkg/crypto"
)

func TestAes(t *testing.T) {
	key := []byte("0123456789abcdef")
	ciphertext, err := crypto.AesEncrypt([]byte("hello AES"), key)
	if err != nil {
		t.Error(err)
		return
	}
	plaintext, err := crypto.AesDecrypt(ciphertext, key)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(plaintext))
}
