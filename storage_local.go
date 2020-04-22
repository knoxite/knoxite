/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io/ioutil"
	"net/url"
	"os"
)

// StorageLocal stores data on the local disk
type StorageLocal struct {
	StorageFilesystem
}

func init() {
	RegisterStorageBackend(&StorageLocal{})
}

// NewBackend returns a StorageLocal backend
func (*StorageLocal) NewBackend(u url.URL) (Backend, error) {
	backend := StorageLocal{}
	storagefs, _ := NewStorageFilesystem(u.Path, &backend)
	backend.StorageFilesystem = storagefs
	return &backend, nil
}

// Location returns the type and location of the repository
func (backend *StorageLocal) Location() string {
	return backend.Path
}

// Close the backend
func (backend *StorageLocal) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend
func (backend *StorageLocal) Protocols() []string {
	return []string{"file"}
}

// Description returns a user-friendly description for this backend
func (backend *StorageLocal) Description() string {
	return "Local File Storage"
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
func (backend StorageLocal) ReadFile(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	return b, err
}

// WriteFile writes a file to disk
func (backend StorageLocal) WriteFile(path string, data []byte) (size uint64, err error) {
	err = ioutil.WriteFile(path, data, 0600)
	return uint64(len(data)), err
}

// DeleteFile deletes a file from disk
func (backend StorageLocal) DeleteFile(path string) error {
	// fmt.Println("Deleting:", path)
	return os.Remove(path)
}
