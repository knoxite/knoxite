/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// A Repository is a collection of backup snapshots.
type Repository struct {
	Version uint      `json:"version"`
	Volumes []*Volume `json:"volumes"`
	Paths   []string  `json:"storage"`
	Key     string    `json:"key"` // key for encrypting data stored with knoxite
	// Owner   string    `json:"owner"`

	backend  BackendManager
	password string // password for knoxite repository file
}

// Const declarations.
const (
	RepositoryVersion   = 4
	repositoryKeyLength = 32 // A random key of 32 is considered safe right now and may be increased later
)

// Error declarations.
var (
	ErrRepositoryIncompatible  = errors.New("the repository is not compatible with this version of Knoxite")
	ErrOpenRepositoryFailed    = errors.New("wrong password or corrupted repository")
	ErrVolumeNotFound          = errors.New("volume not found")
	ErrSnapshotNotFound        = errors.New("snapshot not found")
	ErrGenerateRandomKeyFailed = errors.New("failed to generate a random encryption key for new repository")
)

// NewRepository returns a new repository.
func NewRepository(path, password string) (Repository, error) {
	log.Info("Creating new repository...")
	key, err := generateRandomKey(repositoryKeyLength)
	if err != nil {
		return Repository{}, ErrGenerateRandomKeyFailed
	}

	repository := Repository{
		Version:  RepositoryVersion,
		password: password,
		Key:      key,
	}

	backend, err := BackendFromURL(path)
	if err != nil {
		log.Fatal("Couldn't get backend from URL")
		return repository, err
	}
	repository.backend.AddBackend(&backend)

	err = repository.init()
	return repository, err
}

// generateRandomKey generates a random key with a specific length.
func generateRandomKey(length int) (string, error) {
	log.Debug("Generating random key for repository encryption")
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// OpenRepository opens an existing repository and migrates it if possible.
func OpenRepository(path, password string) (Repository, error) {
	log.Debug("Opening repository...")
	repository := Repository{
		password: password,
	}

	backend, err := BackendFromURL(path)
	if err != nil {
		log.Fatal("Couldn't get backend from URL")
		return repository, err
	}
	b, err := backend.LoadRepository()
	if err != nil {
		log.Fatal("Couldn't load repository")
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
		// migrate to current version
		err = repository.Migrate()
		if err != nil {
			log.Fatalf("Couldn't migrate repository")
			return repository, err
		}
	}

	for _, url := range repository.Paths {
		backend, err := BackendFromURL(url)
		if err != nil {
			log.Fatalf("Couldn't get backend from URL")
			return repository, err
		}
		repository.backend.AddBackend(&backend)
	}

	return repository, err
}

// AddVolume adds a volume to a repository.
func (r *Repository) AddVolume(volume *Volume) error {
	log.Infof("Adding volume %s to repository...", volume.ID)
	r.Volumes = append(r.Volumes, volume)
	return nil
}

// RemoveVolume removes a volume from a repository.
func (r *Repository) RemoveVolume(volume *Volume) error {
	log.Infof("Removing volume %s from repository...", volume.ID)
	for i, v := range r.Volumes {
		if v == volume {
			r.Volumes = append(r.Volumes[:i], r.Volumes[i+1:]...)
			return nil
		}
	}
	return ErrVolumeNotFound
}

// FindVolume finds a volume within a repository.
func (r *Repository) FindVolume(id string) (*Volume, error) {
	log.Debugf("Finding volume %s...", id)
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

// FindSnapshot finds a snapshot within a repository.
func (r *Repository) FindSnapshot(id string) (*Volume, *Snapshot, error) {
	log.Debugf("Finding snapshot %s...", id)
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

// IsEmpty returns true if there a no snapshots stored in a repository.
func (r *Repository) IsEmpty() bool {
	for _, volume := range r.Volumes {
		if len(volume.Snapshots) > 0 {
			return false
		}
	}

	return true
}

// BackendManager returns the repository's BackendManager.
func (r *Repository) BackendManager() *BackendManager {
	return &r.backend
}

// Init creates a new repository.
func (r *Repository) init() error {
	err := r.backend.InitRepository()
	if err != nil {
		log.Fatal("Couldn't initialize repository")
		return err
	}

	return r.Save()
}

// Save writes a repository's metadata.
func (r *Repository) Save() error {
	log.Infof("Saving repository...")
	r.Paths = r.backend.Locations()

	pipe, err := NewEncodingPipeline(CompressionNone, EncryptionAES, r.password)
	if err != nil {
		return err
	}
	b, err := pipe.Encode(r)
	if err != nil {
		return err
	}
	err = r.backend.SaveRepository(b)
	if err == nil {
		log.Info("Saved repository successfully")
	}
	return err
}

// Changes password of repository.
func (r *Repository) ChangePassword(newPassword string) error {
	log.Info("Changing password...")
	r.password = newPassword

	return r.Save()
}

// Migrates a repository to the current version, if possible.
func (r *Repository) Migrate() error {
	log.Infof("Trying to migrate from repository version %s to version %s", r.Version, RepositoryVersion)
	switch v := r.Version; {
	case v < 3:
		return ErrRepositoryIncompatible
	case v == 3:
		log.Infof("Migrating from repository version %s to version 4", v)
		// since the introduction of the repo passwd command there are two keys:
		// - Key is for encryption of the data and will be stored in encrypted repo file
		// - password is for the encryption of the repository (which holds Key)
		// to migrate we need to use the existing repository password as key
		if r.Key == "" {
			r.Key = r.password
			r.Version = 4

			err := r.Save()
			if err == nil {
				log.Infof("Migrated to repository verion 4 successfully")
			}
			return err
		}
	}
	return ErrRepositoryIncompatible
}
