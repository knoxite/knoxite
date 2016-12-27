/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import "net/url"

// StorageDrive stores data on a remote Google Drive
type StorageDrive struct {
	url url.URL
}

// NewStorageDrive returns a StorageDrive object
func NewStorageDrive(u url.URL) *StorageDrive {
	return &StorageDrive{}
}

// Location returns the type and location of the repository
func (backend *StorageDrive) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageDrive) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageDrive) Protocols() []string {
	return []string{"Google Drive"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageDrive) Description() string {
	return "Google Drive Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageDrive) AvailableSpace() (uint64, error) {
	return 0, ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from Google Drive
func (backend *StorageDrive) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	return &[]byte{}, ErrChunkNotFound
}

// StoreChunk stores a single Chunk on Google Drive
func (backend *StorageDrive) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	return 0, ErrStoreChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageDrive) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageDrive) SaveSnapshot(id string, data []byte) error {
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageDrive) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageDrive) LoadRepository() ([]byte, error) {
	return []byte{}, ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository
func (backend *StorageDrive) SaveRepository(data []byte) error {
	return ErrStoreRepositoryFailed
}
