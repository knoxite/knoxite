/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// ChunkError records an error and the index
// that caused it.
type ChunkError struct {
	ChunkNum uint
}

func (e *ChunkError) Error() string {
	return fmt.Sprintf("Could not find chunk #%d", e.ChunkNum)
}

// SeekError records an error and the offset
// that caused it.
type SeekError struct {
	Offset int
}

func (e *SeekError) Error() string {
	return fmt.Sprintf("Could not seek to offset %d", e.Offset)
}

// CheckSumError records an error and the calculated
// checksums that did not match.
type CheckSumError struct {
	Method           string
	ExpectedCheckSum string
	FoundCheckSum    string
}

func (e *CheckSumError) Error() string {
	return fmt.Sprintf("%s mismatch, expected %s, got %s", e.Method, e.ExpectedCheckSum, e.FoundCheckSum)
}

// DataReconstructionError records an error and the associated
// parity information
type DataReconstructionError struct {
	Chunk          Chunk
	BlocksFound    uint
	FailedBackends uint
}

func (e *DataReconstructionError) Error() string {
	return fmt.Sprintf("Could not reconstruct data, got %d out of %d chunks (%d backends missing data)", e.BlocksFound, e.Chunk.DataParts, e.FailedBackends)
}

// DecodeSnapshot restores an entire snapshot to dst
func DecodeSnapshot(repository Repository, snapshot *Snapshot, dst string) (prog chan Progress, err error) {
	prog = make(chan Progress)
	go func() {
		for _, arc := range snapshot.Archives {
			path := filepath.Join(dst, arc.Path)
			err := DecodeArchive(prog, repository, arc, path)
			if err != nil {
				p := newProgressError(err)
				prog <- p
				break
			}
		}
		close(prog)
	}()

	return prog, nil
}

func decodeChunk(repository Repository, chunk Chunk, r io.Reader) (io.ReadCloser, error) {
	var ro io.ReadCloser
	var err error
	if chunk.Encrypted == EncryptionAES {
		ro, err = Decrypt(r, repository.Password)
		if err != nil {
			return nil, err
		}
		defer ro.Close()
	}

	if chunk.Compressed == CompressionGZip {
		ro, err = Uncompress(ro)
		if err != nil {
			return nil, err
		}
		defer ro.Close()
	}

	b, err := ioutil.ReadAll(ro)
	if err != nil {
		return nil, err
	}

	shadata := sha256.Sum256(b)
	shasum := hex.EncodeToString(shadata[:])
	if chunk.DecryptedShaSum != shasum {
		return nil, &CheckSumError{"sha256", chunk.DecryptedShaSum, shasum}
	}

	ro = ioutil.NopCloser(bytes.NewBuffer(b))
	return ro, nil
}

// DecodeArchive restores a single archive to path
func DecodeArchive(progress chan Progress, repository Repository, arc Archive, path string) error {
	p := newProgress(&arc)

	if arc.Type == Directory {
		//fmt.Printf("Creating directory %s\n", path)
		os.MkdirAll(path, arc.Mode)
		p.TotalStatistics.Dirs++
		progress <- p
	} else if arc.Type == SymLink {
		//fmt.Printf("Creating symlink %s -> %s\n", path, arc.PointsTo)
		os.Symlink(arc.PointsTo, path)
		p.TotalStatistics.SymLinks++
		progress <- p
	} else if arc.Type == File {
		parts := uint(len(arc.Chunks))
		//fmt.Printf("Creating file %s (%d chunks).\n", path, parts)

		p.TotalStatistics.Files++
		p.TotalStatistics.Size = arc.Size
		p.TotalStatistics.StorageSize = arc.StorageSize
		progress <- p

		// FIXME: we don't always need to create the path
		// this is just a safety measure for now
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		// write to disk
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, arc.Mode)
		if err != nil {
			return err
		}

		for i := uint(0); i < parts; i++ {
			idx, erri := arc.IndexOfChunk(i)
			if erri != nil {
				return erri
			}

			chunk := arc.Chunks[idx]
			cr, errc := chunk.Load(repository)
			if errc != nil {
				return errc
			}

			r, errc := decodeChunk(repository, chunk, cr)
			if errc != nil {
				return errc
			}
			defer r.Close()

			n, errc := io.Copy(f, r)
			if errc != nil {
				return errc
			}
			cr.Close()

			p.TotalStatistics.Transferred += uint64(n)
			p.CurrentItemStats.Transferred += uint64(n)
			progress <- p
			// fmt.Printf("Chunk OK: %d bytes, sha256: %s\n", size, chunk.DecryptedShaSum)
		}

		f.Sync()
		f.Close()

		// Restore modification time
		err = os.Chtimes(path, arc.ModTime, arc.ModTime)
		if err != nil {
			return err
		}
	}

	// Restore ownerships
	return os.Lchown(path, int(arc.UID), int(arc.GID))
}

var (
	cache map[string][]byte
	mutex = &sync.Mutex{}
)

func init() {
	cache = make(map[string][]byte)
}

func readArchiveChunk(repository Repository, arc Archive, chunkNum uint) (*[]byte, error) {
	var b []byte
	var err error

	idx, err := arc.IndexOfChunk(chunkNum)
	if err != nil {
		return &b, err
	}

	chunk := arc.Chunks[idx]
	mutex.Lock()
	cd, ok := cache[chunk.ShaSum]
	if !ok {
		r, lerr := chunk.Load(repository)
		if lerr != nil {
			return &b, err
		}
		defer r.Close()

		r, err = decodeChunk(repository, chunk, r)
		if err != nil {
			return &b, err
		}
		defer r.Close()

		cd, err = ioutil.ReadAll(r)
		if lerr != nil {
			return &b, err
		}
		cache[chunk.ShaSum] = cd
	}

	mutex.Unlock()
	b = append(b, cd...)

	return &b, nil
}

// ReadArchive reads from an archive
func ReadArchive(repository Repository, arc Archive, offset int, size int) (*[]byte, error) {
	var b []byte

	// fmt.Println("Read req:", offset, size)
	if arc.Type == File {
		neededPart, internalOffset, err := arc.ChunkForOffset(offset)
		if err != nil {
			return &b, err
		}

		for len(b) < size {
			if neededPart >= uint(len(arc.Chunks)) {
				return &b, nil
			}
			cd, err := readArchiveChunk(repository, arc, neededPart)
			if err != nil || len(*cd) == 0 {
				//return b, err
				panic(err)
			}

			d := (*cd)[internalOffset:]
			if err != nil || len(d) == 0 {
				//return b, err
				panic(err)
			}
			if len(d)+len(b) > size {
				b = append(b, d[:size-len(b)]...)
			} else {
				b = append(b, d...)
			}

			internalOffset = 0
			neededPart++
		}

		// cache the next block NOW
		go func() {
			readArchiveChunk(repository, arc, neededPart)
		}()
	}

	return &b, nil
}
