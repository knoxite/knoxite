/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2016, Nicolas Martin <penguwingithub@gmail.com>
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"bytes"
	"net/url"
	"path/filepath"

	"io/ioutil"

	"strconv"

	"github.com/stacktic/dropbox"
)

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	url url.URL
	db  dropbox.Dropbox
}

func NewStorageDropbox(url url.URL) *StorageDropbox {
	storageDB := StorageDropbox{
		url: url,
		db:  *dropbox.NewDropbox(),
	}
	storageDB.db.SetAccessToken(url.User.Username())
	return &storageDB
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

// LoadChunk loads a Chunk from dropbox
func (backend *StorageDropbox) LoadChunk(shasum string, part, totalParts uint) (*[]byte, error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	obj, _, err := backend.db.Download(filepath.Join(backend.url.Path, "chunks", fileName), "", 0)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(obj)
	return &data, err
}

// StoreChunk stores a single Chunk from dropbox
func (backend *StorageDropbox) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (size uint64, err error) {
	fileName := shasum + "." + strconv.FormatUint(uint64(part), 10) + "_" + strconv.FormatUint(uint64(totalParts), 10)

	if _, err := backend.db.Metadata(fileName, false, false, "", "", 1); err != nil {
		// Chunk is already stored
		return 0, nil
	}

	_, err = backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(*data)), len(*data), filepath.Join(backend.url.Path, "chunks", fileName), true, "")
	return uint64(len(*data)), err
}

// LoadSnapshot loads a snapshot
func (backend *StorageDropbox) LoadSnapshot(id string) ([]byte, error) {
	// Getting obj as type io.ReadCloser and reading it out in order to get bytes returned
	obj, _, err := backend.db.Download(filepath.Join(backend.url.Path, "snapshots", id), "", 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveSnapshot stores a snapshot
func (backend *StorageDropbox) SaveSnapshot(id string, data []byte) error {
	_, ErrStoreSnapshotFailed := backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(data)), len(data), filepath.Join(backend.url.Path, "snapshots", id), true, "")
	return ErrStoreSnapshotFailed
}

// InitRepository creates a new repository
func (backend *StorageDropbox) InitRepository() error {
	if _, err := backend.db.CreateFolder(backend.url.Path); err != nil {
	}
	if _, err := backend.db.CreateFolder(filepath.Join(backend.url.Path, "snapshots")); err != nil {
	}
	if _, err := backend.db.CreateFolder(filepath.Join(backend.url.Path, "chunks")); err != nil {
	}
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageDropbox) LoadRepository() ([]byte, error) {
	obj, _, err := backend.db.Download(filepath.Join(backend.url.Path, repoFilename), "", 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository
func (backend *StorageDropbox) SaveRepository(data []byte) error {
	_, err := backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(data)), len(data), filepath.Join(backend.url.Path, repoFilename), true, "")
	return err
}
