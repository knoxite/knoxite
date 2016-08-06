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

	// LoadChunk loads a single Chunk
	LoadChunk(shasum string, part uint) (*[]byte, error)
	// StoreChunk stores a single Chunk
	StoreChunk(shasum string, part uint, data *[]byte) (size uint64, err error)

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

// BackendFromURL returns the matching backend for path
func BackendFromURL(path string) (Backend, error) {
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
	case "":
		return &StorageLocal{
			Path: path,
		}, nil
	default:
		return nil, errors.New("Invalid repository url specified")
	}
}
