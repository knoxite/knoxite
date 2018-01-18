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

// Available encryption algos
const (
	EncryptionNone = iota
	EncryptionAES
)

// Error declarations
var (
	ErrInvalidPassword = errors.New("Empty password not permitted")
)

// Encryptor is a pipeline processor that encrypts data
type Encryptor struct {
	Method uint16

	iv    []byte
	block cipher.Block
}

// NewEncryptor returns a newly configured Encryptor
func NewEncryptor(method uint16, password string) (Encryptor, error) {
	e := Encryptor{
		Method: method,
	}
	if method == EncryptionAES {
		if len(password) == 0 {
			return e, ErrInvalidPassword
		}

		key := sha256.Sum256([]byte(password))
		e.iv = key[:aes.BlockSize]

		var err error
		e.block, err = aes.NewCipher(key[:])
		if err != nil {
			return e, err
		}
	}

	return e, nil
}

// Process encrypts the data
func (e Encryptor) Process(data []byte) ([]byte, error) {
	if e.Method == EncryptionNone {
		return data, nil
	}

	b := make([]byte, len(data))
	encrypter := cipher.NewCFBEncrypter(e.block, e.iv)
	encrypter.XORKeyStream(b, data)

	return b, nil
}

// Decryptor is a pipeline processor that decrypts data
type Decryptor struct {
	Method uint16

	iv    []byte
	block cipher.Block
}

// NewDecryptor returns a newly configured Decryptor
func NewDecryptor(method uint16, password string) (Decryptor, error) {
	e := Decryptor{
		Method: method,
	}
	if method == EncryptionAES {
		if len(password) == 0 {
			return e, ErrInvalidPassword
		}

		key := sha256.Sum256([]byte(password))
		e.iv = key[:aes.BlockSize]

		var err error
		e.block, err = aes.NewCipher(key[:])
		if err != nil {
			return e, err
		}
	}

	return e, nil
}

// Process decrypts the data
func (e Decryptor) Process(data []byte) ([]byte, error) {
	if e.Method == EncryptionNone {
		return data, nil
	}

	b := make([]byte, len(data))
	decrypter := cipher.NewCFBDecrypter(e.block, e.iv)
	decrypter.XORKeyStream(b, data)

	return b, nil
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
