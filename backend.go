/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"errors"
	"net/url"
	"strings"
)

// Backend is used to store and access data
type Backend interface {
	// Location returns the type and location of the repository
	Location() string

	// Protocols returns the Protocol Schemes supported by this backend
	Protocols() []string

	// Description returns a user-friendly description for this backend
	Description() string

	// Close the backend
	Close() error

	// AvailableSpace returns the free space in bytes on this backend
	AvailableSpace() (uint64, error)

	// LoadChunk loads a single Chunk
	LoadChunk(shasum string, part, totalParts uint) (*[]byte, error)
	// StoreChunk stores a single Chunk
	StoreChunk(shasum string, part, totalParts uint, data *[]byte) (uint64, error)

	// LoadSnapshot loads a snapshot
	LoadSnapshot(id string) ([]byte, error)
	// SaveSnapshot stores a snapshot
	SaveSnapshot(id string, data []byte) error

	// InitRepository creates a new repository
	InitRepository() error
	// LoadRepository reads the metadata for a repository
	LoadRepository() ([]byte, error)
	// SaveRepository stores the metadata for a repository
	SaveRepository(data []byte) error
}

// Error declarations
var (
	ErrRepositoryExists      = errors.New("Repository seems to already exist")
	ErrInvalidRepositoryURL  = errors.New("Invalid repository url specified")
	ErrAvailableSpaceUnknown = errors.New("Available space is unknown or undefined")
)

// BackendFromURL returns the matching backend for path
func BackendFromURL(path string) (Backend, error) {
	if strings.Index(path, "://") < 0 {
		path = "file:///" + path
	}

	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http":
		fallthrough
	case "https":
		return &StorageHTTP{
			URL: path,
		}, nil

	case "dropbox":
		return NewStorageDropbox(*u), nil

	case "backblaze":
		return NewStorageBackblaze(*u)

	case "ftp":
		return NewStorageFTP(*u)

	case "s3":
		fallthrough
	case "s3s":
		return NewStorageAmazonS3(*u)

	case "file":
		return NewStorageLocal(path[8:])

	default:
		return nil, ErrInvalidRepositoryURL
	}
}
