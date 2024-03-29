// Package encryption source https://github.com/spotlight21c/aesencryptor
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

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

func EncryptAES(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
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

// AesEncrypt 兼容js的CryptoJS的AES
func AesEncrypt(origData, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	origData = PKCS5Padding(origData, len(key))
	blockMode := cipher.NewCBCEncrypter(block, key) //iv=key
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)

	return hex.EncodeToString(encrypted), nil
}

// AesDecrypt 兼容js的CryptoJS的AES
func AesDecrypt(encrypted, key []byte) ([]byte, error) {
	encrypted, _ = hex.DecodeString(string(encrypted))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

/*
js参考：
function encrypt(data, keyStr) {
 const key = CryptoJS.enc.Utf8.parse(keyStr);
 const encrypted = CryptoJS.AES.encrypt(data, key, {
   mode: CryptoJS.mode.CBC,
   padding: CryptoJS.pad.Pkcs7,
   iv: key
 });
 return encrypted.ciphertext.toString(CryptoJS.enc.Hex);
}

const p = 'test'; // 明文
const k = 'e2b61f44ee172fc9'; // 密钥，长度为16字节（128位）
var encrypted = encrypt(p,k);
console.log("加密后的结果：", encrypted.toString());
*/
