/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"compress/gzip"
	"io"
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
func Uncompress(r io.Reader) (io.ReadCloser, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return zr, nil
}
