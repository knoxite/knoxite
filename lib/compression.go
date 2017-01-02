/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// Which compression algo
const (
	CompressionNone = iota
	CompressionGZip
	CompressionLZW
	CompressionFlate
	CompressionZlib
)

// Compress data
func Compress(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	b = buf.Bytes()

	return b, nil
}

// Uncompress data
func Uncompress(b []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return []byte{}, err
	}
	defer zr.Close()

	return ioutil.ReadAll(zr)
}
