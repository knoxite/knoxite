/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

// Available compression algos
const (
	CompressionNone = iota
	CompressionGZip
	CompressionLZMA
	CompressionLZW
	CompressionFlate
	CompressionZlib
)

// Compress data
func Compress(b []byte, compressionType uint16) ([]byte, error) {
	var buf bytes.Buffer
	var w io.WriteCloser
	var err error

	switch compressionType {
	case CompressionGZip:
		w = gzip.NewWriter(&buf)
	case CompressionLZMA:
		w, err = xz.NewWriter(&buf)
	}
	if err != nil {
		return []byte{}, err
	}
	w.Write(b)
	w.Close()

	b = buf.Bytes()
	return b, nil
}

// Uncompress data
func Uncompress(b []byte, compressionType uint16) ([]byte, error) {
	var zr io.ReadCloser
	var err error

	switch compressionType {
	case CompressionGZip:
		zr, err = gzip.NewReader(bytes.NewReader(b))
	case CompressionLZMA:
		zri, erri := xz.NewReader(bytes.NewReader(b))
		zr = ioutil.NopCloser(zri)
		err = erri
	}
	if err != nil {
		return []byte{}, err
	}
	defer zr.Close()

	return ioutil.ReadAll(zr)
}
