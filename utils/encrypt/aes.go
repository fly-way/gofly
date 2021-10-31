package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// CBC 模式 PKCS7 填充
// key 的 16 位, 24, 32 分别对应 AES-128, AES-192, or AES-256
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}


// AesDynaIvEncrypt, AesDynaIvDecrypt
// 初始化向量与密钥相比有不同的安全性需求，因此IV通常无须保密，然而在大多数情况中，不应当在使用同一密钥的情况下两次使用同一个IV。对于CBC和CFB，
// 重用IV会导致泄露明文首个块的某些信息，亦包括两个不同消息中相同的前缀。对于OFB和CTR而言，重用IV会导致完全失去安全性。
// 另外，在CBC模式中，IV在加密时必须是无法预测的；特别的，在许多实现中使用的产生IV的方法，
// 例如SSL2.0使用的，即采用上一个消息的最后一块密文作为下一个消息的IV，是不安全的。
func AesDynaIvEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	rawData = PKCS7Padding(rawData, blockSize)

	cipherText := make([]byte, blockSize+len(rawData))
	iv := cipherText[:blockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)
	return cipherText, nil
}

func AesDynaIvDecrypt(encryptData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(encryptData) < blockSize {
		return nil, errors.New("cipher text too short")
	}

	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]

	// CBC mode always works in whole blocks.
	if len(encryptData)%blockSize != 0 || len(encryptData) == 0 {
		return nil, errors.New("cipher text is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encryptData, encryptData)
	length := len(encryptData)
	if int(encryptData[length-1]) > length {
		return nil, errors.New("encryptData[length-1] > length")
	}

	encryptData = PKCS7UnPadding(encryptData)
	return encryptData, nil
}


// 固定 iv 加密算法, 可被预测
func AesEncrypt(rawData, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	rawData = PKCS7Padding(rawData, blockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(rawData))
	mode.CryptBlocks(cipherText, rawData)

	return cipherText, nil
}

// 固定 iv 解密算法, 可被预测
func AesDecrypt(encryptData, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(encryptData) < blockSize {
		return encryptData, errors.New("cipher text too short")
	}

	if len(encryptData)%blockSize != 0 || len(encryptData) == 0 {
		return encryptData, errors.New("cipher text is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(encryptData))
	mode.CryptBlocks(origData, encryptData)
	length := len(origData)
	if int(origData[length-1]) > length {
		return encryptData, errors.New("origData[length-1] > length")
	}
	origData = PKCS7UnPadding(origData)

	return origData, nil
}
