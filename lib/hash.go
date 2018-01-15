/*
 * knoxite
 *     Copyright (c) 2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/minio/highwayhash"
)

// Available hash algos
const (
	HashSha256 = iota
	HashHighway256
)

var hashkey [32]byte

// Hash data
func Hash(b []byte, hashtype uint8) string {
	var data [32]byte

	switch hashtype {
	case HashSha256:
		data = sha256.Sum256(b)
	case HashHighway256:
		data = highwayhash.Sum(b, hashkey[:])
	}

	return hex.EncodeToString(data[:])
}
