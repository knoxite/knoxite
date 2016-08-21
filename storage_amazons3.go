/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"net/url"
)

// StorageAmazonS3 stores data on a remote AmazonS3
type StorageAmazonS3 struct {
	url url.URL
}

// NewStorageAmazonS3 returns a StorageAmazonS3 object.
func NewStorageAmazonS3(URL url.URL) (*StorageAmazonS3, error) {
	return &StorageAmazonS3{url: URL}, nil
}

// Location returns the type and location of the repository
func (backend *StorageAmazonS3) Location() string {
	return backend.url.String()
}

// Close the backend
func (backend *StorageAmazonS3) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageAmazonS3) Protocols() []string {
	return []string{"s3", "s3s"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageAmazonS3) Description() string {
	return "Amazon S3 Storage"
}

// LoadChunk loads a Chunk from network
func (backend *StorageAmazonS3) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	return &[]byte{}, ErrChunkNotFound
}

// StoreChunk stores a single Chunk on network
func (backend *StorageAmazonS3) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	return 0, ErrStoreChunkFailed
}

// LoadSnapshot loads a snapshot
func (backend *StorageAmazonS3) LoadSnapshot(id string) ([]byte, error) {
	return []byte{}, ErrSnapshotNotFound
}

// SaveSnapshot stores a snapshot
func (backend *StorageAmazonS3) SaveSnapshot(id string, data []byte) error {
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageAmazonS3) InitRepository() error {
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageAmazonS3) LoadRepository() ([]byte, error) {
	return []byte{}, ErrLoadRepositoryFailed
}

// SaveRepository stores the metadata for a repository
func (backend *StorageAmazonS3) SaveRepository(data []byte) error {
	return ErrStoreRepositoryFailed
}
