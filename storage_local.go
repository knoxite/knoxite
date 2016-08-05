/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	repoFilename = "repository.knox"
)

// StorageLocal stores data on the local disk
type StorageLocal struct {
	Path string
	//	repository Repository
}

// Location returns the type and location of the repository
func (backend *StorageLocal) Location() string {
	return backend.Path
}

// Close the backend
func (backend *StorageLocal) Close() error {
	return nil
}

// Protocol Scheme supported by this backend
func (backend *StorageLocal) Protocol() string {
	return ""
}

// Description returns a user-friendly description for this backend
func (backend *StorageLocal) Description() string {
	return "Local File Storage"
}

// LoadChunk loads a Chunk from disk
func (backend *StorageLocal) LoadChunk(chunk Chunk) ([]byte, error) {
	fileName := filepath.Join(backend.Path, "chunks", chunk.ShaSum)
	b := []byte{}
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	return b, err
}

// StoreChunk stores a single Chunk on disk
func (backend *StorageLocal) StoreChunk(chunk Chunk) (size uint64, err error) {
	fileName := filepath.Join(backend.Path, "chunks", chunk.ShaSum)
	if _, err = os.Stat(fileName); err == nil {
		// Chunk is already stored
		return 0, nil
	}

	err = ioutil.WriteFile(fileName, *chunk.Data, 0600)
	if err != nil {
		fmt.Println(err)
	}
	return uint64(len(*chunk.Data)), err
}

// LoadSnapshot loads a snapshot
func (backend *StorageLocal) LoadSnapshot(id string) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath.Join(backend.Path, "snapshots", id))
	if err != nil {
		fmt.Println(err)
	}

	return b, err
}

// SaveSnapshot stores a snapshot
func (backend *StorageLocal) SaveSnapshot(id string, b []byte) error {
	return ioutil.WriteFile(filepath.Join(backend.Path, "snapshots", id), b, 0600)
}

// InitRepository creates a new repository
func (backend *StorageLocal) InitRepository() error {
	fileName := filepath.Join(backend.Path, repoFilename)
	if _, err := os.Stat(fileName); err == nil {
		// Repo seems to already exist
		return errors.New("Repository seems to already exist")
	}

	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageLocal) LoadRepository() ([]byte, error) {
	b, err := ioutil.ReadFile(filepath.Join(backend.Path, repoFilename))
	if err != nil {
		fmt.Println(err)
	}

	return b, err
}

// SaveRepository stores the metadata for a repository
func (backend *StorageLocal) SaveRepository(b []byte) error {
	fileName := filepath.Join(backend.Path, repoFilename)
	err := ioutil.WriteFile(fileName, b, 0600)
	if err == nil {
		reqPaths := []string{"chunks", "snapshots"}
		for _, reqPath := range reqPaths {
			path := filepath.Join(backend.Path, reqPath)
			if stat, serr := os.Stat(path); serr == nil {
				if !stat.IsDir() {
					return errors.New("Repository contains an invalid file named " + reqPath)
				}
			} else {
				err = os.Mkdir(path, 0700)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}
