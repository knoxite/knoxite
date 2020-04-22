/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

// BackendFactory is used to initialize a new backend
type BackendFactory interface {
	NewBackend(url url.URL) (Backend, error)
	Protocols() []string
}

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
	LoadChunk(shasum string, part, totalParts uint) ([]byte, error)
	// StoreChunk stores a single Chunk
	StoreChunk(shasum string, part, totalParts uint, data []byte) (uint64, error)
	// DeleteChunk deletes a single Chunk
	DeleteChunk(shasum string, part, totalParts uint) error

	// LoadSnapshot loads a snapshot
	LoadSnapshot(id string) ([]byte, error)
	// SaveSnapshot stores a snapshot
	SaveSnapshot(id string, data []byte) error

	// LoadChunkIndex loads the chunk-index
	LoadChunkIndex() ([]byte, error)
	// SaveChunkIndex stores the chunk-index
	SaveChunkIndex(data []byte) error

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
	ErrInvalidUsername       = errors.New("Username wrong or missing")

	backends = []BackendFactory{}
)

// RegisterStorageBackend needs to be called by storage backends to register themselves
func RegisterStorageBackend(factory BackendFactory) {
	backends = append(backends, factory)
}

func newBackendFromProtocol(url url.URL) (Backend, error) {
	for _, backend := range backends {
		for _, p := range backend.Protocols() {
			if p == url.Scheme {
				return backend.NewBackend(url)
			}
		}
	}

	return nil, ErrInvalidRepositoryURL
}

// BackendFromURL returns the matching backend for path
func BackendFromURL(path string) (Backend, error) {
	if !strings.Contains(path, "://") {
		if !filepath.IsAbs(path) {
			var err error
			path, err = filepath.Abs(path)
			if err != nil {
				return nil, err
			}
		}
		path = "file://" + path
	}

	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	return newBackendFromProtocol(*u)
}
