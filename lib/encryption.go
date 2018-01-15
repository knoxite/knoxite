/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
)

// Which encryption algo
const (
	EncryptionNone = iota
	EncryptionAES
)

// Error declarations
var (
	ErrInvalidPassword = errors.New("Empty password not permitted")
)

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
func Encrypt(b []byte, password string) ([]byte, error) {
	var err error
	if len(password) == 0 {
		return []byte{}, ErrInvalidPassword
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Encrypt
	encrypted := make([]byte, len(b))
	err = encryptAESCFB(encrypted, b, key[:], iv)

	return encrypted, err
}

// Decrypt data
func Decrypt(b []byte, password string) ([]byte, error) {
	var err error
	if len(password) == 0 {
		return []byte{}, ErrInvalidPassword
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Decrypt
	decrypted := make([]byte, len(b))
	err = decryptAESCFB(decrypted, b, key[:], iv)

	return decrypted, err
}
