/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/klauspost/reedsolomon"
)

// DecodeSnapshot restores an entire snapshot to dst
func DecodeSnapshot(repository Repository, snapshot Snapshot, dst string) (prog chan Progress, err error) {
	prog = make(chan Progress)
	go func() {
		for _, arc := range snapshot.Items {

			path := filepath.Join(dst, arc.Path)
			err := DecodeArchive(prog, repository, arc, path)
			if err != nil {
				panic(err)
			}
		}
		close(prog)
	}()

	return prog, nil
}

func loadChunk(repository Repository, chunk Chunk) ([]byte, error) {
	enc, err := reedsolomon.New(int(chunk.DataParts), int(chunk.ParityParts))
	if err != nil {
		return []byte{}, err
	}
	pars := make([][]byte, chunk.DataParts+chunk.ParityParts)
	parsFound := 0
	parsMissing := 0
	for i := 0; i < int(chunk.DataParts+chunk.ParityParts); i++ {
		var cerr error
		pars[i], cerr = repository.Backend.LoadChunk(chunk, uint(i))
		if cerr != nil {
			pars[i] = nil
			parsMissing++
			continue
		}
		parsFound++

		if parsFound >= int(chunk.DataParts) {
			var b bytes.Buffer
			bufWriter := bufio.NewWriter(&b)

			if parsMissing > 0 {
				err = enc.Reconstruct(pars)
				if err != nil {
					continue
				}
			}
			err = enc.Join(bufWriter, pars, chunk.Size)
			if err != nil {
				continue
			}
			bufWriter.Flush()
			finalData := b.Bytes()

			if chunk.Encrypted == EncryptionAES {
				data, err := Decrypt(finalData, repository.Password)
				if err != nil {
					return []byte{}, err
				}

				finalData = data
			}

			if chunk.Compressed == CompressionGZip {
				reader := bytes.NewReader(finalData)
				zipreader, err := gzip.NewReader(reader)
				if err != nil {
					return []byte{}, err
				}
				defer zipreader.Close()
				finalData, err = ioutil.ReadAll(zipreader)
				if err != nil {
					return []byte{}, err
				}
			}

			shasumdata := sha256.Sum256(finalData)
			shasum := hex.EncodeToString(shasumdata[:])

			if chunk.DecryptedShaSum != shasum {
				return []byte{}, fmt.Errorf("sha256 mismatch, expected %s got %s", chunk.DecryptedShaSum, shasum)
			}

			return finalData, nil
		}
	}

	return []byte{}, errors.New("Could not reconstruct data")
}

// DecodeArchive restores a single archive to path
func DecodeArchive(progress chan Progress, repository Repository, arc ItemData, path string) error {
	prog := Progress{}
	prog.Path = arc.Path

	if arc.Type == Directory {
		//fmt.Printf("Creating directory %s\n", path)
		os.MkdirAll(path, arc.Mode)
		prog.Statistics.Dirs++
	} else if arc.Type == SymLink {
		//fmt.Printf("Creating symlink %s -> %s\n", path, arc.PointsTo)
		os.Symlink(arc.PointsTo, path)
		prog.Statistics.SymLinks++
	} else if arc.Type == File {
		prog.Statistics.StorageSize = arc.StorageSize
		prog.StorageSize = arc.StorageSize
		parts := len(arc.Chunks)
		//fmt.Printf("Creating file %s (%d chunks).\n", path, parts)

		// write to disk
		os.MkdirAll(filepath.Dir(path), 0755)
		f, ferr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, arc.Mode)
		if ferr != nil {
			return ferr
		}

		for i := int(0); i < parts; i++ {
			for _, chunk := range arc.Chunks {
				if int(chunk.Num) == i {
					finalData, cerr := loadChunk(repository, chunk)
					if cerr != nil {
						return cerr
					}

					// write/save buffer to disk
					_, ferr := f.Write(finalData)
					if ferr != nil {
						return ferr
					}

					prog.Statistics.Size += uint64(len(finalData))
					prog.Size += uint64(len(finalData))
					progress <- prog
					// fmt.Printf("Chunk OK: %d bytes, sha256: %s\n", size, chunk.DecryptedShaSum)
				}
			}
		}

		f.Sync()
		f.Close()
		prog.Statistics.Files++
		// fmt.Printf("Done: %d bytes total\n", totalSize)

		// Restore modification time
		err := os.Chtimes(path, arc.ModTime, arc.ModTime)
		if err != nil {
			return err
		}
	}

	// Restore ownerships
	err := os.Lchown(path, int(arc.UID), int(arc.GID))

	return err
}

var (
	cache map[string][]byte
	mutex = &sync.Mutex{}
)

func init() {
	cache = make(map[string][]byte)

}

// DecodeArchiveData returns the content of a single archive
func DecodeArchiveData(repository Repository, arc ItemData) (dat []byte, stats Stat, err error) {
	if arc.Type == File {
		parts := len(arc.Chunks)

		for i := int(0); i < parts; i++ {
			for _, chunk := range arc.Chunks {
				if int(chunk.Num) == i {
					mutex.Lock()
					cacheData, ok := cache[chunk.ShaSum]
					if ok {
						fmt.Println("Using cached chunk", chunk.ShaSum)
						dat = append(dat, cacheData...)
						mutex.Unlock()
						continue
					}

					finalData, cerr := loadChunk(repository, chunk)
					if cerr != nil {
						return dat, stats, cerr
					}

					stats.StorageSize += uint64(len(finalData))
					stats.Size += uint64(len(finalData))

					dat = append(dat, finalData...)
					cache[chunk.ShaSum] = finalData
					mutex.Unlock()
				}
			}
		}

		stats.Files++
		// fmt.Printf("Done: %d bytes total\n", totalSize)
	}

	return dat, stats, nil
}

func readArchiveChunk(repository Repository, arc ItemData, chunkNum uint64) (dat *[]byte, err error) {
	dat = &[]byte{}
	for _, chunk := range arc.Chunks {
		if chunk.Num == chunkNum {
			mutex.Lock()
			cacheData, ok := cache[chunk.ShaSum]
			if ok {
				//				fmt.Println("Using cached chunk", chunk.ShaSum)
				*dat = append(*dat, cacheData...)
				mutex.Unlock()
				continue
			}

			finalData, err := loadChunk(repository, chunk)
			if err != nil {
				return dat, err
			}

			*dat = append(*dat, finalData...)
			cache[chunk.ShaSum] = finalData
			mutex.Unlock()
		}
	}

	return dat, nil
}

// ReadArchive reads from an archive
func ReadArchive(repository Repository, arc ItemData, offset int64, size int) (dat *[]byte, err error) {
	dat = &[]byte{}
	//	fmt.Println("Read req:", offset, size)
	if arc.Type == File {
		const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

		// calculate needed part for offset
		neededPart := uint64(float64(offset) / float64(fileChunk))
		internalOffset := offset % fileChunk

		for len(*dat) < size {
			b, err := readArchiveChunk(repository, arc, neededPart)
			if err != nil || len(*b) == 0 {
				//return dat, err
				panic(err)
			}

			d := *b
			d = d[internalOffset:]
			if err != nil || len(d) == 0 {
				//return dat, err
				panic(err)
			}
			if len(d) >= size {
				*dat = append(*dat, d[:size]...)
			} else {
				*dat = append(*dat, d...)
			}

			internalOffset = 0
			neededPart++
		}

		// cache the next block NOW
		go func() {
			readArchiveChunk(repository, arc, neededPart)
		}()
	}

	return dat, nil
}
