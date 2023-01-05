package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

// AesEncrypt key 长度 16, 24, 32 分别对应 AES-128, AES-192, AES-256
func AesEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	iv := key[:blockSize]

	plaintext = PKCS7Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv) // CBC
	ciphertext := make([]byte, len(plaintext))
	blockMode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func AesDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	iv := key[:blockSize]

	blockMode := cipher.NewCBCDecrypter(block, iv) // CBC
	plaintext := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(plaintext, ciphertext)
	plaintext, err = PKCS7UnPadding(plaintext, blockSize)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
