/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/knoxite/knoxite/cmd/knoxite/utils"
	"golang.org/x/crypto/scrypt"
)

// EncryptedHeaderPrefix is added to the encrypted configuration to make it
// possible to detect it's an encrypted configuration file.
const EncryptedHeaderPrefix = "knoxiteconf+"

var (
	password string
)

// AESBackend symmetrically encrypts the configuration file using AES-GCM.
type AESBackend struct{}

// NewAESBackend creates the backend.
//
// Given the password is required to encrypt/decrypt the configuration, if the
// URL passed doesn't have a password or PasswordEnvVar is not defined,
// it'll return an error.
func NewAESBackend(u *url.URL) (*AESBackend, error) {
	if _, err := getPassword(u); err != nil {
		return nil, err
	}

	return &AESBackend{}, nil
}

// IsEncrypted returns true and no error if the configuration is encrypted
//
// If the error returned is not nil, an error was returned while opening or
// reading the file.
func IsEncrypted(u *url.URL) (bool, error) {
	path, err := url.QueryUnescape(u.Path)
	if err != nil {
		return false, err
	}
	u.Path = path

	f, err := os.Open(u.Path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	b := make([]byte, 12)
	_, err = f.Read(b)
	if err != nil {
		return false, err
	}

	if string(b) != EncryptedHeaderPrefix {
		return false, nil
	}

	return true, nil
}

func (b *AESBackend) Type() int {
	return CryptoConf
}

// Load configuration file from the given URL and decrypt it.
func (b *AESBackend) Load(u *url.URL) (*Config, error) {
	path, err := url.PathUnescape(u.Path)
	if err != nil {
		return nil, err
	}
	u.Path = path

	config := &Config{url: u}

	if !exist(u.Path) {
		return config, nil
	}

	ciphertext, err := ioutil.ReadFile(u.Path)
	if err != nil {
		return nil, err
	}
	ftype := ciphertext[0:len(EncryptedHeaderPrefix)]
	if string(ftype) != EncryptedHeaderPrefix {
		return nil, errors.New("encrypted configuration header not valid")
	}

	p, err := getPassword(u)
	if err != nil {
		return nil, err
	}

	plaintext, err := decrypt(ciphertext[len(EncryptedHeaderPrefix):], []byte(p))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(plaintext, config)
	if err != nil {
		return nil, err
	}

	config.backend = b
	config.url = u

	return config, nil
}

// Save encrypts then saves the configuration.
func (b *AESBackend) Save(config *Config) error {
	u := config.URL()
	path, err := url.QueryUnescape(u.Path)
	if err != nil {
		return err
	}
	u.Path = path

	cfgDir := filepath.Dir(u.Path)
	if !exist(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return err
		}
	}

	j, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	p, err := getPassword(u)
	if err != nil {
		return err
	}

	ciphertext, err := encrypt(j, []byte(p))
	if err != nil {
		return err
	}

	marked := []byte(EncryptedHeaderPrefix)
	err = ioutil.WriteFile(u.Path, append(marked, ciphertext...), 0644)

	return err
}

func encrypt(data, key []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

func decrypt(data, key []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]
	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

// getPassword checks if we've already read out the password for the
// configuration file from the url or the command line.
func getPassword(u *url.URL) (string, error) {
	var err error

	// check if the password has been changed in the url
	if u.User.Username() != "" && password != u.User.Username() {
		password = u.User.Username()
	}

	if password == "" {
		if u.User.Username() != "" {
			password = u.User.Username()
		} else {
			password, err = utils.ReadPassword("Please type in the password for your configuration file")
		}
	}
	return password, err
}
