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
	CompressionFlate
	CompressionZlib
	CompressionZstd
)

// Compressor is a pipeline processor that compresses data
type Compressor struct {
	Method uint16
}

// Process compresses the data
func (c Compressor) Process(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var w io.WriteCloser
	var err error

	switch c.Method {
	case CompressionNone:
		return data, nil
	case CompressionGZip:
		w = gzip.NewWriter(&buf)
	case CompressionLZMA:
		w, err = xz.NewWriter(&buf)
	}
	if err != nil {
		return []byte{}, err
	}
	w.Write(data)
	w.Close()

	return buf.Bytes(), nil
}

// Decompressor is a pipeline processor that decompresses data
type Decompressor struct {
	Method uint16
}

// Process decompresses the data
func (c Decompressor) Process(data []byte) ([]byte, error) {
	var zr io.ReadCloser
	var err error

	switch c.Method {
	case CompressionNone:
		return data, nil
	case CompressionGZip:
		zr, err = gzip.NewReader(bytes.NewReader(data))
	case CompressionLZMA:
		zri, erri := xz.NewReader(bytes.NewReader(data))
		zr = ioutil.NopCloser(zri)
		err = erri
	}
	if err != nil {
		return []byte{}, err
	}
	defer zr.Close()

	return ioutil.ReadAll(zr)
}
