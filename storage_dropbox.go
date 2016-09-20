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
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/stacktic/dropbox"
)

// StorageDropbox stores data on a remote Dropbox
type StorageDropbox struct {
	url            url.URL
	chunkPath      string
	snapshotPath   string
	repositoryPath string
	db             dropbox.Dropbox
}

// NewStorageDropbox returns a StorageDropbox object
func NewStorageDropbox(url url.URL) *StorageDropbox {
	storageDB := StorageDropbox{
		url:            url,
		chunkPath:      filepath.Join(url.Path, "chunks"),
		snapshotPath:   filepath.Join(url.Path, "snapshots"),
		repositoryPath: filepath.Join(url.Path, repoFilename),
		db:             *dropbox.NewDropbox(),
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
	path := filepath.Join(backend.chunkPath, SubDirForChunk(shasum))
	fileName := filepath.Join(path, shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))

	obj, _, err := backend.db.Download(fileName, "", 0)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(obj)
	return &data, err
}

// StoreChunk stores a single Chunk on dropbox
func (backend *StorageDropbox) StoreChunk(shasum string, part, totalParts uint, data *[]byte) (uint64, error) {
	path := filepath.Join(backend.chunkPath, SubDirForChunk(shasum))
	if _, err := backend.db.CreateFolder(path); err != nil {
		return 0, err
	}
	fileName := filepath.Join(path, shasum+"."+strconv.FormatUint(uint64(part), 10)+"_"+strconv.FormatUint(uint64(totalParts), 10))

	if entry, err := backend.db.Metadata(fileName, false, false, "", "", 1); err == nil {
		// Chunk is already stored
		if int(entry.Bytes) == len(*data) {
			return 0, nil
		}
	}

	//FIXME: this doesn't really chunk anything - it always picks the full data block's size
	entry, err := backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(*data)), len(*data), fileName, true, "")
	return uint64(entry.Bytes), err
}

// LoadSnapshot loads a snapshot
func (backend *StorageDropbox) LoadSnapshot(id string) ([]byte, error) {
	path := filepath.Join(backend.snapshotPath, id)
	// Getting obj as type io.ReadCloser and reading it out in order to get bytes returned
	obj, _, err := backend.db.Download(path, "", 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveSnapshot stores a snapshot
func (backend *StorageDropbox) SaveSnapshot(id string, data []byte) error {
	path := filepath.Join(backend.snapshotPath, id)
	_, err := backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(data)), len(data), path, true, "")
	return err
}

// InitRepository creates a new repository
func (backend *StorageDropbox) InitRepository() error {
	if _, err := backend.db.CreateFolder(backend.url.Path); err != nil {
		return ErrRepositoryExists
	}
	if _, err := backend.db.CreateFolder(backend.snapshotPath); err != nil {
		return ErrRepositoryExists
	}
	if _, err := backend.db.CreateFolder(backend.chunkPath); err != nil {
		return ErrRepositoryExists
	}
	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *StorageDropbox) LoadRepository() ([]byte, error) {
	obj, _, err := backend.db.Download(backend.repositoryPath, "", 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(obj)
}

// SaveRepository stores the metadata for a repository
func (backend *StorageDropbox) SaveRepository(data []byte) error {
	_, err := backend.db.UploadByChunk(ioutil.NopCloser(bytes.NewReader(data)), len(data), backend.repositoryPath, true, "")
	return err
}
