/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"io/ioutil"
	"os"
	"syscall"
)

// StorageLocal stores data on the local disk
type StorageLocal struct {
	StorageFilesystem
}

// NewStorageLocal returns a StorageLocal object
func NewStorageLocal(path string) (*StorageLocal, error) {
	storage := StorageLocal{}
	storagefs, _ := NewStorageFilesystem(path, &storage)
	storage.StorageFilesystem = storagefs
	return &storage, nil
}

// Location returns the type and location of the repository
func (backend *StorageLocal) Location() string {
	return backend.path
}

// Close the backend
func (backend *StorageLocal) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageLocal) Protocols() []string {
	return []string{""}
}

// Description returns a user-friendly description for this backend
func (backend *StorageLocal) Description() string {
	return "Local File Storage"
}

// AvailableSpace returns the free space on this backend
func (backend *StorageLocal) AvailableSpace() (uint64, error) {
	//FIXME: make this cross-platform compatible
	var stat syscall.Statfs_t
	syscall.Statfs(backend.path, &stat)

	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize), nil
}

// CreatePath creates a dir including all its parents dirs, when required
func (backend *StorageLocal) CreatePath(path string) error {
	return os.MkdirAll(path, 0700)
}

// Stat stats a file on disk
func (backend StorageLocal) Stat(path string) (uint64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return uint64(stat.Size()), err
}

// ReadFile reads a file from disk
func (backend StorageLocal) ReadFile(path string) (*[]byte, error) {
	b, err := ioutil.ReadFile(path)
	return &b, err
}

// WriteFile writes a file to disk
func (backend StorageLocal) WriteFile(path string, data *[]byte) (size uint64, err error) {
	err = ioutil.WriteFile(path, *data, 0600)
	return uint64(len(*data)), err
}
