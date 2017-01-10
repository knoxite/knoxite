/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package googledrive

import (
	"net/url"

	knoxite "github.com/knoxite/knoxite/lib"
)

// StorageGoogleDrive stores data on a remote Google Drive
type StorageGoogleDrive struct {
	url url.URL
}

func init() {
	knoxite.RegisterBackendFactory(&StorageGoogleDrive{})
}

// NewBackend returns a StorageGoogleDrive backend
func (*StorageGoogleDrive) NewBackend(u url.URL) (knoxite.Backend, error) {
	return &StorageGoogleDrive{}, knoxite.ErrInvalidRepositoryURL
}

// Location returns the type and location of the repository
func (backend *StorageGoogleDrive) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageGoogleDrive) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageGoogleDrive) Protocols() []string {
	return []string{"gdrive"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageGoogleDrive) Description() string {
	return "Google Drive Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageGoogleDrive) AvailableSpace() (uint64, error) {
	return 0, knoxite.ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from Google Drive
func (backend *StorageGoogleDrive) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	return []byte{}, knoxite.ErrLoadChunkFailed
}

// StoreChunk stores a single Chunk on Google Drive
func (backend *StorageGoogleDrive) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	return 0, knoxite.ErrStoreChunkFailed
}

// DeleteChunk deletes a single Chunk
func (backend *StorageGoogleDrive) DeleteChunk(shasum string, parts, totalParts uint) error {
	// FIXME: implement this
	return knoxite.ErrDeleteChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageGoogleDrive) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, knoxite.ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageGoogleDrive) SaveSnapshot(id string, data []byte) error {
	return knoxite.ErrStoreSnapshotFailed
}

// LoadChunkIndex reads the chunk-index
func (backend *StorageGoogleDrive) LoadChunkIndex() ([]byte, error) {
	return []byte{}, knoxite.ErrLoadChunkIndexFailed
}

// SaveChunkIndex stores the chunk-index
func (backend *StorageGoogleDrive) SaveChunkIndex(data []byte) error {
	return knoxite.ErrStoreChunkIndexFailed
}

// InitRepository creates a new repository
func (backend *StorageGoogleDrive) InitRepository() error {
	return knoxite.ErrInvalidRepositoryURL
}

// LoadRepository reads the metadata for a repository
func (backend *StorageGoogleDrive) LoadRepository() ([]byte, error) {
	return []byte{}, knoxite.ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository
func (backend *StorageGoogleDrive) SaveRepository(data []byte) error {
	return knoxite.ErrStoreRepositoryFailed
}
