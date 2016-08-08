/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import uuid "github.com/nu7hatch/gouuid"

// A Volume contains various snapshots
// MUST BE encrypted
type Volume struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Snapshots   []string `json:"snapshots"`
}

// NewVolume creates a new volume
func NewVolume(name, description string) (*Volume, error) {
	vol := Volume{
		Name:        name,
		Description: description,
	}

	u, err := uuid.NewV4()
	if err != nil {
		return &vol, err
	}
	vol.ID = u.String()[:8]

	return &vol, nil
}

// AddSnapshot adds a snapshot to a volume
func (v *Volume) AddSnapshot(id string) error {
	v.Snapshots = append(v.Snapshots, id)
	return nil
}

// LoadSnapshot loads a snapshot from a repository
func (v *Volume) LoadSnapshot(id string, repository *Repository) (Snapshot, error) {
	for _, snapshot := range v.Snapshots {
		if snapshot == id {
			snapshot, err := openSnapshot(id, repository)
			return snapshot, err
		}
	}

	return Snapshot{}, ErrSnapshotNotFound
}
