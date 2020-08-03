/*
 * knoxite
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */

package utils

import (
	"errors"
	"strings"
	"github.com/knoxite/knoxite"
)

var (
	ErrEncryptionUnknown  = errors.New("unknown encryption format")
	ErrCompressionUnknown = errors.New("unknown compression format")
)
// CompressionTypeFromString returns the compression type from a user-specified string
func CompressionTypeFromString(s string) (uint16, error) {
	switch strings.ToLower(s) {
	case "":
		// default is none
		fallthrough
	case "none":
		return knoxite.CompressionNone, nil
	case "flate":
		return knoxite.CompressionFlate, nil
	case "gzip":
		return knoxite.CompressionGZip, nil
	case "lzma":
		return knoxite.CompressionLZMA, nil
	case "zlib":
		return knoxite.CompressionZlib, nil
	case "zstd":
		return knoxite.CompressionZstd, nil
	}

	return 0, ErrCompressionUnknown
}

// CompressionText returns a user-friendly string indicating the compression algo that was used
// returns "unknown" when none is found
func CompressionText(enum int) string {
	switch enum {
	case knoxite.CompressionNone:
		return "none"
	case knoxite.CompressionFlate:
		return "Flate"
	case knoxite.CompressionGZip:
		return "GZip"
	case knoxite.CompressionLZMA:
		return "LZMA"
	case knoxite.CompressionZlib:
		return "zlib"
	case knoxite.CompressionZstd:
		return "zstd"
	}

	return "unknown"
}

// EncryptionTypeFromString returns the encryption type from a user-specified string
func EncryptionTypeFromString(s string) (uint16, error) {
	switch strings.ToLower(s) {
	case "":
		// default is AES
		fallthrough
	case "aes":
		return knoxite.EncryptionAES, nil
	case "none":
		return knoxite.EncryptionNone, nil
	}

	return 0, ErrEncryptionUnknown
}

// EncryptionText returns a user-friendly string indicating the encryption algo that was used
func EncryptionText(enum int) string {
	switch enum {
	case knoxite.EncryptionNone:
		return "none"
	case knoxite.EncryptionAES:
		return "AES"
	}

	return "unknown"
}
