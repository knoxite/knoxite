/*
 * knoxite
 *     Copyright (c) 2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"encoding/gob"
)

// PipelineProcessor is a simple interface to process data.
type PipelineProcessor interface {
	Process(data []byte) ([]byte, error)
}

// Pipeline passes data through various steps (for compression, encryption etc).
type Pipeline struct {
	Processors []PipelineProcessor
}

// NewEncodingPipeline returns a new pipeline consisting of a compressor and an encryptor.
func NewEncodingPipeline(compression, encryption uint16, password string) (Pipeline, error) {
	encryptor, err := NewEncryptor(encryption, password)
	if err != nil {
		return Pipeline{}, err
	}

	return Pipeline{
		Processors: []PipelineProcessor{
			Compressor{
				Method: compression,
			},
			encryptor,
		},
	}, nil
}

// NewDecodingPipeline returns a new pipeline consisting of a decryptor and a decompressor.
func NewDecodingPipeline(compression, encryption uint16, password string) (Pipeline, error) {
	decryptor, err := NewDecryptor(encryption, password)
	if err != nil {
		return Pipeline{}, err
	}

	return Pipeline{
		Processors: []PipelineProcessor{
			decryptor,
			Decompressor{
				Method: compression,
			},
		},
	}, nil
}

// Process sends the data through all configured processors and returns the result.
func (p *Pipeline) Process(data []byte) ([]byte, error) {
	var err error
	for _, proc := range p.Processors {
		data, err = proc.Process(data)
		if err != nil {
			break
		}
	}

	return data, err
}

// Encode gob-encodes an object and sends the data through all configured processors and returns the result.
func (p *Pipeline) Encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return p.Process(buf.Bytes())
}

// Decode sends the data through all configured processors and gob-decodes the result.
func (p *Pipeline) Decode(b []byte, data interface{}) error {
	b, err := p.Process(b)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	return gob.NewDecoder(buf).Decode(data)
}
