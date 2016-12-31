/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"encoding/json"
	"errors"
)

// A Repository is a collection of backup snapshots
// MUST BE encrypted
type Repository struct {
	Version uint      `json:"version"`
	Volumes []*Volume `json:"volumes"`
	Paths   []string  `json:"storage"`
	// Owner   string    `json:"owner"`

	Backend  BackendManager `json:"-"`
	Password string         `json:"-"`
}

// Const declarations
const (
	RepositoryVersion = 1
)

// Error declarations
var (
	ErrOpenRepositoryFailed = errors.New("Wrong password or corrupted repository")
	ErrVolumeNotFound       = errors.New("Volume not found")
	ErrSnapshotNotFound     = errors.New("Snapshot not found")
)

// NewRepository returns a new repository
func NewRepository(path, password string) (Repository, error) {
	repository := Repository{
		Version:  RepositoryVersion,
		Password: password,
	}
	backend, err := BackendFromURL(path)
	if err != nil {
		return repository, err
	}
	repository.Backend.AddBackend(&backend)

	err = repository.init()
	return repository, err
}

// OpenRepository opens an existing repository
func OpenRepository(path, password string) (Repository, error) {
	repository := Repository{
		Password: password,
	}
	backend, err := BackendFromURL(path)
	if err != nil {
		return repository, err
	}

	b, err := backend.LoadRepository()
	if err != nil {
		return repository, err
	}

	b, err = Decrypt(b, password)
	if err == nil {
		err = json.Unmarshal(b, &repository)
	}
	// If decrypt _or_ unmarshal failed, abort
	if err != nil {
		return repository, ErrOpenRepositoryFailed
	}

	for _, url := range repository.Paths {
		backend, berr := BackendFromURL(url)
		if berr != nil {
			return repository, berr
		}
		repository.Backend.AddBackend(&backend)
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

// Init creates a new repository
func (r *Repository) init() error {
	err := r.Backend.InitRepository()
	if err == nil {
		err = r.Save()
	}

	return err
}

// Save writes a repository's metadata
func (r *Repository) Save() error {
	r.Paths = r.Backend.Locations()

	b, err := json.Marshal(*r)
	if err != nil {
		return err
	}

	b, err = Encrypt(b, r.Password)
	if err == nil {
		err = r.Backend.SaveRepository(b)
	}
	return err
}
