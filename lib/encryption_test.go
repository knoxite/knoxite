/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestEncryption(t *testing.T) {
	testPassword := "this_is_a_password"
	b := []byte("1234567890")

	be, err := Encrypt(b, testPassword)
	if err != nil {
		t.Error(err)
	}

	dr, err := Decrypt(bytes.NewReader(be), testPassword)
	if err != nil {
		t.Error(err)
	}
	bd, err := ioutil.ReadAll(dr)
	if err != nil {
		t.Error(err)
	}

	if string(b) != string(bd) {
		t.Error("Data mismatch after encryption & decryption cycle.")
	}
}

func TestEmptyPassword(t *testing.T) {
	b := []byte("1234567890")
	_, err := Encrypt(b, "")
	if err != ErrInvalidPassword {
		t.Errorf("Expected %v, got %v", ErrInvalidPassword, err)
	}

	_, err = Decrypt(bytes.NewReader(b), "")
	if err != ErrInvalidPassword {
		t.Errorf("Expected %v, got %v", ErrInvalidPassword, err)
	}
}
