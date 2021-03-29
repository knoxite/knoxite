/*
 * knoxite
 *     Copyright (c) 2021, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021, Nicolas Martin <penguwin@penguwin.eu>
 *     TODO
 *
 *   For license see LICENSE
 */

package alibaba

import (
	"net/url"

	"github.com/knoxite/knoxite"
)

// AlibabaStorage stores data on the alibaba $PRODUCTNAME cloud
type AlibabaStorage struct {
	url url.URL
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterStorageBackend(&AlibabaStorage{})
}

// NewBackend returns a AlibabaStorage backend.
func (*AlibabaStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	// parse the user and password from the url
	if u.User == nil || u.User.Username() == "" {
		return &AlibabaStorage{}, knoxite.ErrInvalidUsername
	}
	pw, pwexist := u.User.Password()
	if !pwexist {
		return &AlibabaStorage{}, knoxite.ErrInvalidPassword
	}

	// TODO: initialize an alibaba api `client` here

	backend := AlibabaStorage{
		url: u,
		// client: client
	}

	fs, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &AlibabaStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

// Location returns the type and location of the repository.
func (backend *AlibabaStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *AlibabaStorage) Close() error {
	// NOTE: You may not need to close anything for alibaba
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *AlibabaStorage) Protocols() []string {
	return []string{"Alibaba"}
}

// Description returns a user-friendly description for this backend.
func (backend *AlibabaStorage) Description() string {
	return "Alibaba Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *AlibabaStorage) AvailableSpace() (uint64, error) {
	// TODO: return the available space on alibaba cloud
	return 0, nil
}

// CreatePath creates a dir including all its parent dirs, when required.
func (backend *AlibabaStorage) CreatePath(path string) error {
	// TODO: create the given path on alibaba cloud
	return nil
}

// Stat returns the size of a file.
func (backend *AlibabaStorage) Stat(path string) (uint64, error) {
	// TODO: return the size for the file from the given path.
	return 0, nil
}

// ReadFile reads a file from alibaba cloud.
func (backend *AlibabaStorage) ReadFile(path string) ([]byte, error) {
	// TODO: read the file for the given path on alibaba cloud and return the
	// data.
	return nil, nil
}

// WriteFile write files on alibaba cloud.
func (backend *AlibabaStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	// TODO: write the data in the given path on the alibaba cloud
	return 0, nil
}

// DeleteFile deletes a file from alibaba cloud.
func (backend *AlibabaStorage) DeleteFile(path string) error {
	// TODO: delete the file on the given path
	return nil
}
