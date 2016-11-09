/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/stacktic/dropbox"
)

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	url url.URL
	// chunkPath      string
	// snapshotPath   string
	// repositoryPath string
	db *dropbox.Dropbox
	StorageFilesystem
}

// NewStorageDropbox returns a StorageDropbox object
func NewStorageDropbox(u url.URL) (*StorageDropbox, error) {
	storage := StorageDropbox{
		url: u,
		// chunkPath:      filepath.Join(u.Path, "chunks"),
		// snapshotPath:   filepath.Join(u.Path, "snapshots"),
		// repositoryPath: filepath.Join(u.Path, repoFilename),
		db: dropbox.NewDropbox(),
	}

	storageDB, err := NewStorageFilesystem(u.Path, &storage)
	if err != nil {
		return &StorageDropbox{}, err
	}
	storage.StorageFilesystem = storageDB

	ak, _ := base64.StdEncoding.DecodeString("aXF1bGs0a25vajIydGtt")
	as, _ := base64.StdEncoding.DecodeString("N3htbmlhcDV0cmE5NTE5")
	storage.db.SetAppInfo(string(ak), string(as))

	if storage.url.User == nil || len(storage.url.User.Username()) == 0 {
		if err := storage.db.Auth(); err != nil {
			panic(err)
		}
		storage.url.User = url.User(storage.db.AccessToken())
	} else {
		storage.db.SetAccessToken(storage.url.User.Username())
	}

	return &storage, nil
}

// Location returns the type and location of the repository
func (backend *StorageDropbox) Location() string {
	return backend.url.String()
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

// AvailableSpace returns the free space on this backend
func (backend *StorageDropbox) AvailableSpace() (uint64, error) {
	account, err := backend.db.GetAccountInfo()
	if err != nil {
		return 0, err
	}

	return uint64(account.QuotaInfo.Quota - account.QuotaInfo.Shared - account.QuotaInfo.Normal), nil
}

// CreatePath creates a dir including all its parent dirs, when required
func (backend *StorageDropbox) CreatePath(path string) error {
	slicedPath := strings.Split(path, string(filepath.Separator))
	for i := range slicedPath {
		if i == 0 {
			// don't try to create root-dir
			continue
		}
		if _, err := backend.db.CreateFolder(filepath.Join(slicedPath[:i+1]...)); err != nil {
			// We only want to return an error when creating the last directory
			// in this path failed. Parent dirs _may_ already exist
			if i+1 == len(slicedPath) {
				return err
			}
		}
	}
	return nil
}

// Stat returns the size of a file
func (backend *StorageDropbox) Stat(path string) (uint64, error) {
	entry, err := backend.db.Metadata(path, false, false, "", "", 1)
	return uint64(entry.Bytes), err
}

// ReadFile reads a file from dropbox
func (backend *StorageDropbox) ReadFile(path string) (*[]byte, error) {
	obj, _, err := backend.db.Download(path, "", 0)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(obj)
	return &data, err
}

// WriteFile write files on dropbox
func (backend *StorageDropbox) WriteFile(path string, data *[]byte) (size uint64, err error) {
	_, err = backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(*data)), len(*data), path, true, "")
	return uint64(len(*data)), err
}
