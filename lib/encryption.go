/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
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

// Encrypt data
func Encrypt(b []byte, password string) ([]byte, error) {
	var err error
	if len(password) == 0 {
		return []byte{}, ErrInvalidPassword
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Encrypt
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, len(b))
	aesEncrypter := cipher.NewCFBEncrypter(block, iv)
	aesEncrypter.XORKeyStream(encrypted, b)

	return encrypted, err
}

// Decrypt data
func Decrypt(r io.Reader, password string) (io.ReadCloser, error) {
	if len(password) == 0 {
		return nil, ErrInvalidPassword
	}

	var key = sha256.Sum256([]byte(password))
	var iv = key[:aes.BlockSize]

	// Decrypt
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	reader := &cipher.StreamReader{S: stream, R: r}

	return ioutil.NopCloser(reader), nil
}
