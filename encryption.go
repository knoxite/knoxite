/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	// "reflect"
)

// Which encryption algo
const (
	EncryptionNone = iota
	EncryptionAES
)

// EncryptionText returns a user-friendly string indicating the encryption algo that was used
func EncryptionText(enum int) string {
	switch enum {
	case EncryptionNone:
		return "none"
	case EncryptionAES:
		return "AES"
	}

	return "unknown"
}

// encryptAESCFB encryptes src to dst
func encryptAESCFB(dst, src, key, iv []byte) error {
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(dst, src)
	return nil
}

// decryptAESCFB dcryptes src to dst
func decryptAESCFB(dst, src, key, iv []byte) error {
	aesBlockDecrypter, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(dst, src)
	return nil
}

// Encrypt data
func Encrypt(data []byte, password string) ([]byte, error) {
	var err error
	if len(password) == 0 {
		return []byte{}, errors.New("Empty password not permitted")
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Encrypt
	encrypted := make([]byte, len(data))
	err = encryptAESCFB(encrypted, data, key[:], iv)

	// Decryption check
	/*	decrypted := make([]byte, len(data))
		err = DecryptAESCFB(decrypted, encrypted, key[:], iv)
		if err != nil || !reflect.DeepEqual(data, decrypted) {
			panic(err)
		} */

	return encrypted, err
}

// Decrypt data
func Decrypt(data []byte, password string) ([]byte, error) {
	var err error
	if len(password) == 0 {
		return []byte{}, errors.New("Empty password not permitted")
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Decrypt
	decrypted := make([]byte, len(data))
	err = decryptAESCFB(decrypted, data, key[:], iv)

	return decrypted, err
}
