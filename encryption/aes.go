// Package encryption source https://github.com/spotlight21c/aesencryptor
package encryption

import (
	"bytes"
	"crypto/aes"
	"errors"
)

func EncryptAES(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, errors.New("plain content empty")
	}

	content := PKCS5Padding(src, block.BlockSize())
	encrypted := make([]byte, len(content))
	NewECBEncrypter(block).CryptBlocks(encrypted, content)

	return encrypted, nil
}

func DecryptAES(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := NewECBDecrypter(block)
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = PKCS5UnPadding(origData)

	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	pContent := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, pContent...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}
