package miio

import (
	"bytes"
	"fmt"
)

// Pad using PKCS7 padding scheme.
func pkcs7Pad(data []byte, blockSize int) []byte {
	length := len(data)
	padLength := (blockSize - (length % blockSize))
	pad := bytes.Repeat([]byte{byte(padLength)}, padLength)
	return append(data, pad...)
}

// Unpad using PKCS7 padding scheme.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	srcLen := len(data)
	paddingLen := int(data[srcLen-1])
	if paddingLen >= srcLen || paddingLen > blockSize {
		return nil, fmt.Errorf("padding error")
	}
	return data[:srcLen-paddingLen], nil
}
