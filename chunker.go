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
	"math"
	"os"
	"sync"
)

// Which compression algo
const (
	CompressionNone = iota
	CompressionGZip
	CompressionLZW
	CompressionFlate
	CompressionZlib
)

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

// Chunk stores an encrypted chunk alongside with its metadata
// MUST BE encrypted
type Chunk struct {
	Data            *[]byte `json:"-"`
	Size            int     `json:"size"`
	DecryptedShaSum string  `json:"decrypted_sha256"`
	ShaSum          string  `json:"sha256"`
	Encrypted       int     `json:"encrypted"`
	Compressed      int     `json:"compressed"`
	Num             uint64  `json:"num"`
}

type inputChunk struct {
	Data []byte
	Num  uint64
}

func processChunk(id int, compress, encrypt bool, password string, jobs <-chan inputChunk, results chan<- Chunk, wg *sync.WaitGroup) {
	for j := range jobs {
		//		fmt.Println("\tWorker", id, "processing job", j.Num, len(j.Data))

		finalData := j.Data
		if compress {
			var compdata bytes.Buffer
			zipwriter := gzip.NewWriter(&compdata)
			zipwriter.Write(j.Data)
			zipwriter.Close()
			finalData = compdata.Bytes()
		}

		if encrypt {
			encryptedData, err := Encrypt(finalData, password)
			if err != nil {
				panic(err)
			}

			finalData = encryptedData
		}
		shasumdata := sha256.Sum256(finalData)
		shasum := hex.EncodeToString(shasumdata[:])
		decshasumdata := sha256.Sum256(j.Data)
		decshasum := hex.EncodeToString(decshasumdata[:])

		cd := Chunk{
			Data:            &finalData,
			Size:            len(finalData),
			DecryptedShaSum: decshasum,
			ShaSum:          shasum,
			Encrypted:       EncryptionAES,
			Compressed:      CompressionNone,
			Num:             j.Num,
		}
		if compress {
			cd.Compressed = CompressionGZip
		}
		if !encrypt {
			cd.Encrypted = EncryptionNone
		}

		results <- cd
		wg.Done()
	}
}

// chunkFile divides filename into chunks of 1MiB each
func chunkFile(filename string, compress, encrypt bool, password string) (chan Chunk, error) {
	c := make(chan Chunk)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalParts := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	// fmt.Printf("Splitting %s to %d pieces.\n", filename, totalParts)

	wg := &sync.WaitGroup{}
	jobs := make(chan inputChunk)
	for w := 1; w <= 4; w++ {
		go processChunk(w, compress, encrypt, password, jobs, c, wg)
	}

	wg.Add(int(totalParts))
	go func() {
		for i := uint64(0); i < totalParts; i++ {
			partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
			partBuffer := make([]byte, partSize)

			_, err = io.ReadFull(file, partBuffer)
			if err != nil {
				panic(err)
			}
			j := inputChunk{
				Data: partBuffer,
				Num:  i,
			}

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
