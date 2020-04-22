/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"errors"
)

// A Repository is a collection of backup snapshots
// MUST BE encrypted
type Repository struct {
	Version uint      `json:"version"`
	Volumes []*Volume `json:"volumes"`
	Paths   []string  `json:"storage"`
	// Owner   string    `json:"owner"`

	backend  BackendManager
	password string
}

// Const declarations
const (
	RepositoryVersion = 3
)

// Error declarations
var (
	ErrRepositoryIncompatible = errors.New("The repository is not compatible with this version of Knoxite")
	ErrOpenRepositoryFailed   = errors.New("Wrong password or corrupted repository")
	ErrVolumeNotFound         = errors.New("Volume not found")
	ErrSnapshotNotFound       = errors.New("Snapshot not found")
)

// NewRepository returns a new repository
func NewRepository(path, password string) (Repository, error) {
	repository := Repository{
		Version:  RepositoryVersion,
		password: password,
	}

	backend, err := BackendFromURL(path)
	if err != nil {
		return repository, err
	}
	repository.backend.AddBackend(&backend)

	err = repository.init()
	return repository, err
}

// OpenRepository opens an existing repository
func OpenRepository(path, password string) (Repository, error) {
	repository := Repository{
		password: password,
	}

	backend, err := BackendFromURL(path)
	if err != nil {
		return repository, err
	}
	b, err := backend.LoadRepository()
	if err != nil {
		return repository, err
	}

	pipe, err := NewDecodingPipeline(CompressionNone, EncryptionAES, password)
	if err != nil {
		return repository, err
	}
	err = pipe.Decode(b, &repository)
	if err != nil {
		return repository, ErrOpenRepositoryFailed
	}
	if repository.Version < RepositoryVersion {
		return repository, ErrRepositoryIncompatible
	}

	for _, url := range repository.Paths {
		backend, berr := BackendFromURL(url)
		if berr != nil {
			return repository, berr
		}
		repository.backend.AddBackend(&backend)
	}

	return repository, err
}

// AddVolume adds a volume to a repository
func (r *Repository) AddVolume(volume *Volume) error {
	r.Volumes = append(r.Volumes, volume)
	return nil
}

// FindVolume finds a volume within a repository
func (r *Repository) FindVolume(id string) (*Volume, error) {
	if id == "latest" && len(r.Volumes) > 0 {
		return r.Volumes[len(r.Volumes)-1], nil
	}

	for _, volume := range r.Volumes {
		if volume.ID == id {
			return volume, nil
		}
	}

	return &Volume{}, ErrVolumeNotFound
}

// FindSnapshot finds a snapshot within a repository
func (r *Repository) FindSnapshot(id string) (*Volume, *Snapshot, error) {
	if id == "latest" {
		latestVolume := &Volume{}
		latestSnapshot := &Snapshot{}
		found := false
		for _, volume := range r.Volumes {
			for _, snapshotID := range volume.Snapshots {
				snapshot, err := volume.LoadSnapshot(snapshotID, r)
				if err == nil {
					if !found || snapshot.Date.Sub(latestSnapshot.Date) > 0 {
						latestSnapshot = snapshot
						latestVolume = volume
						found = true
					}
				}
			}
		}

		if found {
			return latestVolume, latestSnapshot, nil
		}
	} else {
		for _, volume := range r.Volumes {
			snapshot, err := volume.LoadSnapshot(id, r)
			if err == nil {
				return volume, snapshot, err
			}
		}
	}

	return &Volume{}, &Snapshot{}, ErrSnapshotNotFound
}

// IsEmpty returns true if there a no snapshots stored in a repository
func (r *Repository) IsEmpty() bool {
	for _, volume := range r.Volumes {
		if len(volume.Snapshots) > 0 {
			return false
		}
	}

	return true
}

// BackendManager returns the repository's BackendManager
func (r *Repository) BackendManager() *BackendManager {
	return &r.backend
}

// Init creates a new repository
func (r *Repository) init() error {
	err := r.backend.InitRepository()
	if err != nil {
		return err
	}

	return r.Save()
}

// Save writes a repository's metadata
func (r *Repository) Save() error {
	r.Paths = r.backend.Locations()

	pipe, err := NewEncodingPipeline(CompressionNone, EncryptionAES, r.password)
	if err != nil {
		return err
	}
	b, err := pipe.Encode(r)
	if err != nil {
		return err
	}
	return r.backend.SaveRepository(b)
}
