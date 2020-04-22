/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"testing"
)

func TestEncryption(t *testing.T) {
	testPassword := "this_is_a_password"
	b := []byte("1234567890")

	epipe, err := NewEncodingPipeline(CompressionNone, EncryptionAES, testPassword)
	if err != nil {
		t.Error(err)
	}
	be, err := epipe.Process(b)
	if err != nil {
		t.Error(err)
	}

	dpipe, err := NewDecodingPipeline(CompressionNone, EncryptionAES, testPassword)
	if err != nil {
		t.Error(err)
	}
	bd, err := dpipe.Process(be)
	if err != nil {
		t.Error(err)
	}

	if string(b) != string(bd) {
		t.Error("Data mismatch after encryption & decryption cycle.")
	}
}

func TestEmptyPassword(t *testing.T) {
	_, err := NewEncodingPipeline(CompressionNone, EncryptionAES, "")
	if err != ErrInvalidPassword {
		t.Errorf("Expected %v, got %v", ErrInvalidPassword, err)
	}

	_, err = NewDecodingPipeline(CompressionNone, EncryptionAES, "")
	if err != ErrInvalidPassword {
		t.Errorf("Expected %v, got %v", ErrInvalidPassword, err)
	}
}
