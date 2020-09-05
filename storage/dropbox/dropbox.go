/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE
 */

package dropbox

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"github.com/tj/go-dropbox"
	"github.com/tj/go-dropy"

	"github.com/knoxite/knoxite"
)

// DropboxStorage stores data on a remote Dropbox.
type DropboxStorage struct {
	url   url.URL
	dropy *dropy.Client
	knoxite.StorageFilesystem
}

func init() {
	knoxite.RegisterStorageBackend(&DropboxStorage{})
}

// NewBackend returns a DropboxStorage backend.
func (*DropboxStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	user := u.User.Username()
	if user == "" {
		return &DropboxStorage{}, knoxite.ErrInvalidUsername
	}

	backend := DropboxStorage{
		url:   u,
		dropy: dropy.New(dropbox.New(dropbox.NewConfig(user))),
	}

	fs, err := knoxite.NewStorageFilesystem(u.Path, &backend)
	if err != nil {
		return &DropboxStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

// Location returns the type and location of the repository.
func (backend *DropboxStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *DropboxStorage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *DropboxStorage) Protocols() []string {
	return []string{"dropbox"}
}

// Description returns a user-friendly description for this backend.
func (backend *DropboxStorage) Description() string {
	return "Dropbox Storage"
}

// AvailableSpace returns the free space on this backend.
func (backend *DropboxStorage) AvailableSpace() (uint64, error) {
	space, err := backend.dropy.Client.Users.GetSpaceUsage()
	if err != nil {
		return 0, err
	}
	return space.Allocation.Allocated - space.Allocation.Used, nil
}

// CreatePath creates a dir including all its parent dirs, when required.
func (backend *DropboxStorage) CreatePath(path string) error {
	return backend.dropy.Mkdir(path)
}

// Stat returns the size of a file.
func (backend *DropboxStorage) Stat(path string) (uint64, error) {
	fileinfo, err := backend.dropy.Stat(path)
	if err != nil {
		return 0, err
	}
	return uint64(fileinfo.Size()), nil
}

// ReadFile reads a file from dropbox.
func (backend *DropboxStorage) ReadFile(path string) ([]byte, error) {
	file, err := backend.dropy.Download(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// WriteFile write files on dropbox.
func (backend *DropboxStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	return uint64(len(data)), backend.dropy.Upload(path, bytes.NewReader(data))
}

// DeleteFile deletes a file from dropbox.
func (backend *DropboxStorage) DeleteFile(path string) error {
	return backend.dropy.Delete(path)
}
