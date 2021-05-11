/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io"
	"os"
	"sync"

	"github.com/restic/chunker"
)

const (
	preferredChunkSize = 1 * (1 << 20) // 1 MiB
)

// Chunk stores an encrypted chunk alongside with its metadata.
type Chunk struct {
	Data          *[][]byte `json:"-"`
	DataParts     uint      `json:"data_parts"`
	ParityParts   uint      `json:"parity_parts"`
	OriginalSize  int       `json:"original_size"`
	Size          int       `json:"size"`
	DecryptedHash string    `json:"decrypted_hash"`
	Hash          string    `json:"hash"`
	Num           uint      `json:"num"`
}

// ChunkResult is used to transfer either a chunk or an error down the channel.
type ChunkResult struct {
	Chunk Chunk
	Error error
}

type inputChunk struct {
	Data []byte
	Num  uint
}

func processChunk(password string, opts StoreOptions, jobs <-chan inputChunk, chunks chan<- ChunkResult, wg *sync.WaitGroup) {
	pipe, _ := NewEncodingPipeline(opts.Compress, opts.Encrypt, password)

	for j := range jobs {
		// fmt.Println("\tWorker", id, "processing job", j.Num, len(j.Data))

		b, err := pipe.Process(j.Data)
		if err != nil {
			chunks <- ChunkResult{Error: err}
			wg.Done()
			continue
		}

		hashsum := Hash(b, HashHighway256)
		orighashsum := Hash(j.Data, HashHighway256)

		c := Chunk{
			DataParts:     opts.DataParts,
			ParityParts:   opts.ParityParts,
			OriginalSize:  len(j.Data),
			Size:          len(b),
			DecryptedHash: orighashsum,
			Hash:          hashsum,
			Num:           j.Num,
		}

		if opts.ParityParts > 0 {
			pars, err := redundantData(b, int(opts.DataParts), int(opts.ParityParts))
			if err != nil {
				chunks <- ChunkResult{Error: err}
				wg.Done()
				continue
			}
			c.Data = &pars
		} else {
			c.DataParts = 1
			c.Data = &[][]byte{b}
		}

		chunks <- ChunkResult{Chunk: c}
		wg.Done()
	}
}

// chunkFile divides filename into chunks of 1MiB each.
func chunkFile(filename string, password string, opts StoreOptions) (<-chan ChunkResult, error) {
	c := make(chan ChunkResult)

	file, err := os.Open(filename)
	if err != nil {
		return c, err
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan inputChunk)
	for w := 1; w <= 4; w++ {
		go processChunk(password, opts, jobs, c, wg)
	}

	wg.Add(1)
	go func() {
		chunker := chunker.NewWithBoundaries(file, chunker.Pol(0x3DA3358B4DC173), chunker.MinSize, preferredChunkSize)

		i := uint(0)
		for {
			buf := make([]byte, preferredChunkSize)
			chunk, err := chunker.Next(buf)
			if err == io.EOF {
				wg.Done()
				break
			}
			if err != nil {
				c <- ChunkResult{Error: err}
				wg.Done()
				break
			}

			wg.Add(1)
			j := inputChunk{
				Data: chunk.Data,
				Num:  i,
			}

			i++
			jobs <- j
		}
		_ = file.Close()
	}()

	go func() {
		wg.Wait()
		close(jobs)
		close(c)
	}()

	return c, nil
}
