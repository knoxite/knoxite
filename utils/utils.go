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
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/knoxite/knoxite"
	"github.com/muesli/crunchy"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	ErrPasswordMismatch   = errors.New("Passwords did not match")
	ErrEncryptionUnknown  = errors.New("unknown encryption format")
	ErrCompressionUnknown = errors.New("unknown compression format")
)

func ReadPassword(prompt string) (string, error) {
	var tty io.WriteCloser
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close()
	}

	fmt.Fprint(tty, prompt+" ")
	buf, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(tty)

	return string(buf), err
}

func ReadPasswordTwice(prompt, promptConfirm string) (string, error) {
	pw, err := ReadPassword(prompt)
	if err != nil {
		return pw, err
	}

	crunchErr := crunchy.NewValidator().Check(pw)
	if crunchErr != nil {
		fmt.Printf("Password is considered unsafe: %v\n", crunchErr)
		fmt.Printf("Are you sure you want to use this password (y/N)?: ")
		var buf string
		_, err = fmt.Scan(&buf)
		if err != nil {
			return pw, err
		}

		buf = strings.TrimSpace(buf)
		buf = strings.ToLower(buf)
		if buf != "y" {
			return pw, crunchErr
		}
	}

	pwconfirm, err := ReadPassword(promptConfirm)
	if err != nil {
		return pw, err
	}
	if pw != pwconfirm {
		return pw, ErrPasswordMismatch
	}

	return pw, nil
}

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
