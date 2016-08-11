/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	URL string
}

// Location returns the type and location of the repository
func (backend *StorageDropbox) Location() string {
	return backend.URL
}

// Close the backend
func (backend *StorageDropbox) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageDropbox) Protocols() []string {
	return []string{"dropbox"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageDropbox) Description() string {
	return "Dropbox Storage"
}

// LoadChunk loads a Chunk from network
func (backend *StorageDropbox) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	return &[]byte{}, ErrChunkNotFound
}

// StoreChunk stores a single Chunk on network
func (backend *StorageDropbox) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	return 0, ErrStoreChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageDropbox) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageDropbox) SaveSnapshot(id string, data []byte) error {
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageDropbox) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageDropbox) LoadRepository() ([]byte, error) {
	return []byte{}, ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository
func (backend *StorageDropbox) SaveRepository(data []byte) error {
	return ErrStoreRepositoryFailed
}
