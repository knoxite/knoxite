/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/DataDog/zstd"
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
	case CompressionFlate:
		w, err = flate.NewWriter(&buf, flate.DefaultCompression)
	case CompressionGZip:
		w = gzip.NewWriter(&buf)
	case CompressionLZMA:
		w, err = xz.NewWriter(&buf)
	case CompressionZlib:
		w = zlib.NewWriter(&buf)
	case CompressionZstd:
		w = zstd.NewWriter(&buf)
	}
	if err != nil {
		return []byte{}, err
	}

	n, err := w.Write(data)
	if err != nil {
		return []byte{}, err
	}
	if n != len(data) {
		return []byte{}, fmt.Errorf("Could not write all data to compressor")
	}
	err = w.Close()
	if err != nil {
		return []byte{}, err
	}

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
	case CompressionFlate:
		zr = flate.NewReader(bytes.NewReader(data))
	case CompressionGZip:
		zr, err = gzip.NewReader(bytes.NewReader(data))
	case CompressionLZMA:
		zri, erri := xz.NewReader(bytes.NewReader(data))
		zr = ioutil.NopCloser(zri)
		err = erri
	case CompressionZlib:
		zr, err = zlib.NewReader(bytes.NewReader(data))
	case CompressionZstd:
		zr = zstd.NewReader(bytes.NewReader(data))
	}
	if err != nil {
		return []byte{}, err
	}
	defer zr.Close()

	return ioutil.ReadAll(zr)
}
