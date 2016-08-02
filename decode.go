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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
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
					finalData, cerr := repository.Backend.LoadChunk(chunk)
					if cerr != nil {
						return cerr
					}

					if chunk.Encrypted == EncryptionAES {
						data, err := Decrypt(finalData, repository.Password)
						if err != nil {
							return err
						}

						finalData = data
					}

					if chunk.Compressed == CompressionGZip {
						reader := bytes.NewReader(finalData)
						zipreader, err := gzip.NewReader(reader)
						if err != nil {
							return err
						}
						defer zipreader.Close()
						finalData, err = ioutil.ReadAll(zipreader)
						if err != nil {
							return err
						}
					}

					shasumdata := sha256.Sum256(finalData)
					shasum := hex.EncodeToString(shasumdata[:])
					prog.Statistics.Size += uint64(len(finalData))
					prog.Size += uint64(len(finalData))

					if chunk.DecryptedShaSum != shasum {
						return errors.New("ERROR: sha256 mismatch")
					}

					// write/save buffer to disk
					_, ferr := f.Write(finalData)
					if ferr != nil {
						return ferr
					}
					// fmt.Printf("Chunk OK: %d bytes, sha256: %s\n", size, chunk.DecryptedShaSum)
				}
				progress <- prog
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

// DecodeArchiveData restores a single archive to path
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

					finalData, err := repository.Backend.LoadChunk(chunk)
					origData := finalData
					if err != nil {
						return dat, stats, err
					}

					if chunk.Encrypted == EncryptionAES {
						data, err := Decrypt(finalData, repository.Password)
						if err != nil {
							return dat, stats, err
						}

						finalData = data
					}

					if chunk.Compressed == CompressionGZip {
						reader := bytes.NewReader(finalData)
						zipreader, err := gzip.NewReader(reader)
						if err != nil {
							return dat, stats, err
						}
						defer zipreader.Close()
						finalData, err = ioutil.ReadAll(zipreader)
						if err != nil {
							return dat, stats, err
						}
					}

					shasumdata := sha256.Sum256(finalData)
					shasum := hex.EncodeToString(shasumdata[:])
					stats.StorageSize += uint64(len(origData))
					stats.Size += uint64(len(finalData))

					if chunk.DecryptedShaSum != shasum {
						return dat, stats, errors.New("ERROR: sha256 mismatch")
					}

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

			finalData, err := repository.Backend.LoadChunk(chunk)
			if err != nil {
				return dat, err
			}

			if chunk.Encrypted == EncryptionAES {
				data, err := Decrypt(finalData, repository.Password)
				if err != nil {
					return dat, err
				}

				finalData = data
			}

			if chunk.Compressed == CompressionGZip {
				reader := bytes.NewReader(finalData)
				zipreader, err := gzip.NewReader(reader)
				if err != nil {
					return dat, err
				}
				defer zipreader.Close()
				finalData, err = ioutil.ReadAll(zipreader)
				if err != nil {
					return dat, err
				}
			}

			/*			shasumdata := sha256.Sum256(finalData)
						shasum := hex.EncodeToString(shasumdata[:])

						if chunk.DecryptedShaSum != shasum {
							return dat, errors.New("ERROR: sha256 mismatch")
						}*/

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
