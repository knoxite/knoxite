/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import "net/url"

// StorageBackblaze stores data on a remote Backblaze
type StorageBackblaze struct {
	URL string
}

func NewStorageBackblaze(URL url.URL) (*StorageBackblaze, error) {
	return &StorageBackblaze{}, nil

}

// Location returns the type and location of the repository
func (backend *StorageBackblaze) Location() string {
	return backend.URL
}

// Close the backend
func (backend *StorageBackblaze) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageBackblaze) Protocols() []string {
	return []string{"backblaze"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageBackblaze) Description() string {
	return "Backblaze Storage"
}

func (backend *StorageBackblaze) AvailableSpace() (uint64, error) {
	return 0, ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from backblaze
func (backend *StorageBackblaze) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	return &[]byte{}, ErrChunkNotFound
}

// StoreChunk stores a single Chunk from backblaze
func (backend *StorageBackblaze) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	return 0, ErrStoreChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageBackblaze) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageBackblaze) SaveSnapshot(id string, data []byte) error {
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageBackblaze) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageBackblaze) LoadRepository() ([]byte, error) {
	return []byte{}, ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository
func (backend *StorageBackblaze) SaveRepository(data []byte) error {
	return ErrStoreRepositoryFailed
}
