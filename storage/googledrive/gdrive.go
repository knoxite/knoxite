/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package googledrive

import (
	"net/url"

	"github.com/knoxite/knoxite"
)

// GoogleDriveStorage stores data on a remote Google Drive.
type GoogleDriveStorage struct {
	url url.URL
}

func init() {
	knoxite.RegisterStorageBackend(&GoogleDriveStorage{})
}

// NewBackend returns a GoogleDriveStorage backend.
func (*GoogleDriveStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	return &GoogleDriveStorage{}, knoxite.ErrInvalidRepositoryURL
}

// Location returns the type and location of the repository.
func (backend *GoogleDriveStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *GoogleDriveStorage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *GoogleDriveStorage) Protocols() []string {
	return []string{"gdrive"}
}

// Description returns a user-friendly description for this backend.
func (backend *GoogleDriveStorage) Description() string {
	return "Google Drive Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *GoogleDriveStorage) AvailableSpace() (uint64, error) {
	return 0, knoxite.ErrAvailableSpaceUnknown
}

// LoadChunk loads a Chunk from Google Drive.
func (backend *GoogleDriveStorage) LoadChunk(shasum string, part, totalParts uint) ([]byte, error) {
	return []byte{}, knoxite.ErrLoadChunkFailed
}

// StoreChunk stores a single Chunk on Google Drive.
func (backend *GoogleDriveStorage) StoreChunk(shasum string, part, totalParts uint, data []byte) (size uint64, err error) {
	return 0, knoxite.ErrStoreChunkFailed
}

// DeleteChunk deletes a single Chunk.
func (backend *GoogleDriveStorage) DeleteChunk(shasum string, parts, totalParts uint) error {
	// FIXME: implement this
	return knoxite.ErrDeleteChunkFailed
}

// LoadSnapshot loads a snapshot.
func (backend *GoogleDriveStorage) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, knoxite.ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot.
func (backend *GoogleDriveStorage) SaveSnapshot(id string, data []byte) error {
	return knoxite.ErrStoreSnapshotFailed
}

// LoadChunkIndex reads the chunk-index.
func (backend *GoogleDriveStorage) LoadChunkIndex() ([]byte, error) {
	return []byte{}, knoxite.ErrLoadChunkIndexFailed
}

// SaveChunkIndex stores the chunk-index.
func (backend *GoogleDriveStorage) SaveChunkIndex(data []byte) error {
	return knoxite.ErrStoreChunkIndexFailed
}

// InitRepository creates a new repository.
func (backend *GoogleDriveStorage) InitRepository() error {
	return knoxite.ErrInvalidRepositoryURL
}

// LoadRepository reads the metadata for a repository.
func (backend *GoogleDriveStorage) LoadRepository() ([]byte, error) {
	return []byte{}, knoxite.ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository.
func (backend *GoogleDriveStorage) SaveRepository(data []byte) error {
	return knoxite.ErrStoreRepositoryFailed
}

// LockRepository locks the repository and prevents other instances from
// concurrent access.
func (backend *GoogleDriveStorage) LockRepository(data []byte) ([]byte, error) {
	// TODO: implement
	return nil, nil
}

// UnlockRepository releases the lock.
func (backend *GoogleDriveStorage) UnlockRepository() error {
	// TODO: implement
	return nil
}
