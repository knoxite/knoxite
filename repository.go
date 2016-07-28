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
	"fmt"
)

// A Repository is a collection of backup snapshots
// MUST BE encrypted
type Repository struct {
	//	Owner   string    `json:"owner"`
	Volumes []*Volume `json:"volumes"`

	Path     string `json:"-"`
	Password string `json:"-"`

	Backend Backend
}

// NewRepository returns a new repository
func NewRepository(path, password string) (Repository, error) {
	repository := Repository{
		//		Owner:    owner,
		Path:     path,
		Password: password,
	}
	backend, err := BackendFromURL(path)
	if err != nil {
		return repository, err
	}
	repository.Backend = backend
	fmt.Printf("Using backend: %s\n", backend.Description())

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
	repository.Backend = backend
	fmt.Printf("Using backend: %s\n", backend.Description())

	b, err := repository.Backend.LoadRepository()

	decb, err := Decrypt(b, password)
	if err == nil {
		err = json.Unmarshal(decb, &repository)
	}
	return repository, err
}

// AddVolume adds a volume to a repository
func (r *Repository) AddVolume(volume *Volume) error {
	r.Volumes = append(r.Volumes, volume)
	return nil
}

// FindSnapshot finds a snapshot within a repository
func (r *Repository) FindSnapshot(id string) (*Volume, *Snapshot, error) {
	for _, volume := range r.Volumes {
		snapshot, err := volume.LoadSnapshot(id, r)
		if err == nil {
			return volume, &snapshot, err
		}
	}

	return &Volume{}, &Snapshot{}, errors.New("Snapshot not found")
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
	//	b, err := json.MarshalIndent(*r, "", "    ")
	b, err := json.Marshal(*r)
	if err != nil {
		return err
	}
	//	fmt.Printf("Repository created: %s\n", string(b))

	encb, err := Encrypt(b, r.Password)
	if err == nil {
		err = r.Backend.SaveRepository(encb)
	}
	return err
}
