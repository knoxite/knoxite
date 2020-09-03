/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"fmt"
	"path/filepath"
	"strconv"
)

const (
	// RepoFilename is the default filename for the repository data.
	RepoFilename = "repository.knoxite"
	// ChunkIndexFilename is the default filename for the chunk-index.
	ChunkIndexFilename = "index"
	chunksDirname      = "chunks"
	snapshotsDirname   = "snapshots"
	// LockFilename is the default filename for repository locks.
	LockFilename = "repository.lock"
)

// BackendFilesystem is used to store and access data on a filesytem based backend.
type BackendFilesystem interface {
	// Stat stats a file on disk
	Stat(path string) (uint64, error)
	// CreatePath creates a dir including all its parents dirs, when required
	CreatePath(path string) error
	// ReadFile reads a file from disk
	ReadFile(path string) ([]byte, error)
	// WriteFile writes a file to disk
	WriteFile(path string, data []byte) (uint64, error)
	// DeleteFile deletes a file from disk
	DeleteFile(path string) error
}

// StorageFilesystem is bridging a BackendFilesystem to a Backend interface.
type StorageFilesystem struct {
	Path           string
	chunkPath      string
	snapshotPath   string
	chunkIndexPath string
	repositoryPath string
	lockPath       string

	storage *BackendFilesystem
}

// NewStorageFilesystem returns a StorageFilesystem object.
func NewStorageFilesystem(path string, storage BackendFilesystem) (StorageFilesystem, error) {
	s := StorageFilesystem{
		Path:           path,
		chunkPath:      filepath.Join(path, chunksDirname),
		snapshotPath:   filepath.Join(path, snapshotsDirname),
		chunkIndexPath: filepath.Join(path, chunksDirname, ChunkIndexFilename),
		repositoryPath: filepath.Join(path, RepoFilename),
		lockPath:       filepath.Join(path, LockFilename),
		storage:        &storage,
	}
	return s, nil
}

// LoadChunk loads a Chunk from disk.
func (backend StorageFilesystem) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	path := filepath.Join(backend.chunkPath, SubDirForChunk(shasum))
	fileName := filepath.Join(path, shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))

	return (*backend.storage).ReadFile(fileName)
}

// StoreChunk stores a single Chunk on disk.
func (backend StorageFilesystem) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	path := filepath.Join(backend.chunkPath, SubDirForChunk(shasum))
	fileName := filepath.Join(path, shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))

	n, err := (*backend.storage).Stat(fileName)
	if err == nil && n == uint64(len(data)) {
		return 0, nil
	}

	err = (*backend.storage).CreatePath(path)
	if err != nil {
		return 0, err
	}

	return (*backend.storage).WriteFile(fileName, data)
}

// DeleteChunk deletes a single Chunk.
func (backend StorageFilesystem) DeleteChunk(shasum string, part, totalParts uint) error {
	path := filepath.Join(backend.chunkPath, SubDirForChunk(shasum))
	fileName := filepath.Join(path, shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))

	return (*backend.storage).DeleteFile(fileName)
}

// LoadSnapshot loads a snapshot.
func (backend StorageFilesystem) LoadSnapshot(id string) ([]byte, error) {
	b, err := (*backend.storage).ReadFile(filepath.Join(backend.snapshotPath, id))
	if err != nil {
		fmt.Println(err)
	}

	return b, err
}

// SaveSnapshot stores a snapshot.
func (backend StorageFilesystem) SaveSnapshot(id string, b []byte) error {
	_, err := (*backend.storage).WriteFile(filepath.Join(backend.snapshotPath, id), b)
	return err
}

// LoadChunkIndex reads the chunk-index.
func (backend StorageFilesystem) LoadChunkIndex() ([]byte, error) {
	b, err := (*backend.storage).ReadFile(backend.chunkIndexPath)
	if err != nil {
		return []byte{}, err
	}
	return b, err
}

// SaveChunkIndex stores the chunk-index.
func (backend StorageFilesystem) SaveChunkIndex(b []byte) error {
	_, err := (*backend.storage).WriteFile(backend.chunkIndexPath, b)
	return err
}

// InitRepository creates a new repository.
func (backend StorageFilesystem) InitRepository() error {
	if _, err := (*backend.storage).Stat(backend.repositoryPath); err == nil {
		// Repo seems to already exist
		return ErrRepositoryExists
	}
	paths := []string{backend.chunkPath, backend.snapshotPath}
	for _, path := range paths {
		if _, serr := (*backend.storage).Stat(path); serr == nil {
			return ErrRepositoryExists
			/*
				if !stat.IsDir() {
					return &os.PathError{Op: "create", Path: path, Err: errors.New("Repository path contains an invalid file")}
				}
			*/
		}
		err := (*backend.storage).CreatePath(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadRepository reads the metadata for a repository.
func (backend StorageFilesystem) LoadRepository() ([]byte, error) {
	b, err := (*backend.storage).ReadFile(backend.repositoryPath)
	if err != nil {
		return []byte{}, err
	}

	return b, err
}

// SaveRepository stores the metadata for a repository.
func (backend StorageFilesystem) SaveRepository(b []byte) error {
	_, err := (*backend.storage).WriteFile(backend.repositoryPath, b)
	return err
}

// LockRepository locks the repository and prevents other instances from
// concurrent access.
func (backend StorageFilesystem) LockRepository() error {
	b, err := (*backend.storage).ReadFile(backend.lockPath)
	if err == nil {
		return ErrRepositoryLocked
	}

	_, err = (*backend.storage).WriteFile(backend.lockPath, b)
	return err
}

// UnlockRepository releases the lock.
func (backend StorageFilesystem) UnlockRepository() error {
	return (*backend.storage).DeleteFile(backend.lockPath)
}

// SubDirForChunk files a chunk into a subdir, based on the chunks name.
func SubDirForChunk(id string) string {
	return filepath.Join(id[0:2], id[2:4])
}
