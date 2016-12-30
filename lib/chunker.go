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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/muesli/chunker"
)

// Which compression algo
const (
	CompressionNone = iota
	CompressionGZip
	CompressionLZW
	CompressionFlate
	CompressionZlib

	preferredChunkSize = 1 * (1 << 20) // 1 MiB
)

// Chunk stores an encrypted chunk alongside with its metadata
// MUST BE encrypted
type Chunk struct {
	Data            *[][]byte `json:"-"`
	DataParts       uint      `json:"data_parts"`
	ParityParts     uint      `json:"parity_parts"`
	OriginalSize    int       `json:"original_size"`
	Size            int       `json:"size"`
	DecryptedShaSum string    `json:"decrypted_sha256"`
	ShaSum          string    `json:"sha256"`
	Encrypted       int       `json:"encrypted"`
	Compressed      int       `json:"compressed"`
	Num             uint      `json:"num"`
}

type inputChunk struct {
	Data []byte
	Num  uint
}

func processChunk(id int, compress, encrypt bool, password string, dataParts, parityParts int, jobs <-chan inputChunk, chunks chan<- Chunk, wg *sync.WaitGroup) {
	for j := range jobs {
		// fmt.Println("\tWorker", id, "processing job", j.Num, len(j.Data))

		var err error
		b := j.Data
		if compress {
			var cb bytes.Buffer
			w := gzip.NewWriter(&cb)
			w.Write(b)
			w.Close()
			b = cb.Bytes()
		}
		if encrypt {
			b, err = Encrypt(b, password)
			if err != nil {
				panic(err)
			}
		}

		shadata := sha256.Sum256(b)
		shasum := hex.EncodeToString(shadata[:])
		shadata = sha256.Sum256(j.Data)
		origshasum := hex.EncodeToString(shadata[:])

		c := Chunk{
			DataParts:       uint(dataParts),
			ParityParts:     uint(parityParts),
			OriginalSize:    len(j.Data),
			Size:            len(b),
			DecryptedShaSum: origshasum,
			ShaSum:          shasum,
			Encrypted:       EncryptionAES,
			Compressed:      CompressionNone,
			Num:             j.Num,
		}
		if compress {
			c.Compressed = CompressionGZip
		}
		if !encrypt {
			c.Encrypted = EncryptionNone
		}
		if parityParts > 0 {
			pars, err := redundantData(b, dataParts, parityParts)
			if err != nil {
				panic(err)
			}
			c.Data = &pars
		} else {
			c.DataParts = 1
			c.Data = &[][]byte{b}
		}

		chunks <- c
		wg.Done()
	}
}

// chunkFile divides filename into chunks of 1MiB each
func chunkFile(filename string, compress, encrypt bool, password string, dataParts, parityParts int) (chan Chunk, error) {
	c := make(chan Chunk)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan inputChunk)
	for w := 1; w <= 4; w++ {
		go processChunk(w, compress, encrypt, password, dataParts, parityParts, jobs, c, wg)
	}

	wg.Add(1)
	go func() {
		chunker := chunker.NewWithBoundaries(file, chunker.Pol(0x3DA3358B4DC173), chunker.MinSize, preferredChunkSize)

		i := uint(0)
		for {
			partBuffer := make([]byte, preferredChunkSize)
			chunk, err := chunker.Next(partBuffer)
			if err == io.EOF {
				wg.Done()
				break
			}
			if err != nil {
				panic(err)
			}

			wg.Add(1)
			j := inputChunk{
				Data: chunk.Data,
				Num:  i,
			}

			i++
			jobs <- j
		}
		file.Close()
	}()

	go func() {
		wg.Wait()
		close(jobs)
		close(c)
	}()

	return c, nil
}

// CompressionText returns a user-friendly string indicating the compression algo that was used
func CompressionText(enum int) string {
	switch enum {
	case CompressionNone:
		return "none"
	case CompressionGZip:
		return "GZip"
	case CompressionLZW:
		return "LZW"
	case CompressionFlate:
		return "Flate"
	case CompressionZlib:
		return "zlib"
	}

	return "unknown"
}
