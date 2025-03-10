package miio

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
)

func descryptPayload(encryptedPayload, key, iv []byte) (decrypted []byte, err error) {
	if len(encryptedPayload) == 0 {
		return decrypted, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCBCDecrypter(block, iv)

	decrypted = make([]byte, len(encryptedPayload))
	stream.CryptBlocks(decrypted, encryptedPayload)

	decrypted, err = pkcs7Unpad(decrypted, block.BlockSize())
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func encryptPayload(payload, key, iv []byte) (encrypted []byte, err error) {
	if payload == nil {
		return encrypted, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	payload = pkcs7Pad(payload, block.BlockSize())
	stream := cipher.NewCBCEncrypter(block, iv)

	encrypted = make([]byte, len(payload))
	stream.CryptBlocks(encrypted, payload)

	return encrypted, nil
}

func getKeyAndIV(token []byte) (key []byte, iv []byte, err error) {
	hash := md5.New()
	_, err = hash.Write(token)
	if err != nil {
		return nil, nil, err
	}
	key = hash.Sum(nil)

	hash = md5.New()
	_, err = hash.Write(key)
	if err != nil {
		return nil, nil, err
	}
	_, err = hash.Write(token)
	if err != nil {
		return nil, nil, err
	}
	iv = hash.Sum(nil)

	return
}
