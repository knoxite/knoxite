/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
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

// RemoveSnapshot removes a snapshot from a volume
func (v *Volume) RemoveSnapshot(id string) error {
	snapshots := []string{}
	found := false

	for _, snapshot := range v.Snapshots {
		if snapshot == id {
			found = true
		} else {
			snapshots = append(snapshots, snapshot)
		}
	}

	if !found {
		return ErrSnapshotNotFound
	}

	v.Snapshots = snapshots
	return nil
}

// LoadSnapshot loads a snapshot within a volume from a repository
func (v *Volume) LoadSnapshot(id string, repository *Repository) (*Snapshot, error) {
	for _, snapshot := range v.Snapshots {
		if snapshot == id {
			snapshot, err := openSnapshot(id, repository)
			return snapshot, err
		}
	}

	return &Snapshot{}, ErrSnapshotNotFound
}
