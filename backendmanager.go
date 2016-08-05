/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import "errors"

// BackendManager storfes data on multiple backends
type BackendManager struct {
	Backends []*Backend
}

// AddBackend adds a backend
func (backend *BackendManager) AddBackend(be *Backend) {
	backend.Backends = append(backend.Backends, be)
}

// Locations returns the urls for all backends
func (backend *BackendManager) Locations() []string {
	paths := []string{}
	for _, be := range backend.Backends {
		paths = append(paths, (*be).Location())
	}

	return paths
}

// LoadChunk loads a Chunk from backends
func (backend *BackendManager) LoadChunk(chunk Chunk) ([]byte, error) {
	for _, be := range backend.Backends {
		b, err := (*be).LoadChunk(chunk)
		if err == nil {
			return b, err
		}
	}

	return []byte{}, errors.New("Unable to load chunk from any storage backend")
}

// StoreChunk stores a single Chunk on backends
func (backend *BackendManager) StoreChunk(chunk Chunk) (size uint64, err error) {
	for _, be := range backend.Backends {
		_, err := (*be).StoreChunk(chunk)
		if err != nil {
			return 0, err
		}
	}

	return uint64(chunk.Size), nil
}

// LoadSnapshot loads a snapshot
func (backend *BackendManager) LoadSnapshot(id string) ([]byte, error) {
	for _, be := range backend.Backends {
		b, err := (*be).LoadSnapshot(id)
		if err == nil {
			return b, err
		}
	}

	return []byte{}, errors.New("Unable to load snapshot from any storage backend")
}

// SaveSnapshot stores a snapshot on all storage backends
func (backend *BackendManager) SaveSnapshot(id string, b []byte) error {
	for _, be := range backend.Backends {
		err := (*be).SaveSnapshot(id, b)
		if err != nil {
			return err
		}
	}

	return nil
}

// InitRepository creates a new repository
func (backend *BackendManager) InitRepository() error {
	for _, be := range backend.Backends {
		err := (*be).InitRepository()
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadRepository reads the metadata for a repository
func (backend *BackendManager) LoadRepository() ([]byte, error) {
	for _, be := range backend.Backends {
		b, err := (*be).LoadRepository()
		if err == nil {
			return b, err
		}
	}

	return []byte{}, errors.New("Unable to load repository from any storage backend")
}

// SaveRepository stores the metadata for a repository
func (backend *BackendManager) SaveRepository(b []byte) error {
	for _, be := range backend.Backends {
		err := (*be).SaveRepository(b)
		if err != nil {
			return err
		}
	}

	return nil
}
