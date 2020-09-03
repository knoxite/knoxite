/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestVolumeCreate(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	vol, verr := NewVolume("test_name", "test_description")
	{
		r, err := NewRepository(dir, testPassword)
		if err != nil {
			t.Errorf("Failed creating repository: %s", err)
			return
		}

		if verr == nil {
			verr = r.AddVolume(vol)
			if verr != nil {
				t.Errorf("Failed creating volume: %s", verr)
				return
			}

			serr := r.Save()
			if serr != nil {
				t.Errorf("Failed saving volume: %s", serr)
				return
			}
		}

		if r.Close() != nil {
			t.Errorf("Failed closing repository: %s", err)
			return
		}
	}

	{
		r, err := OpenRepository(dir, testPassword, false)
		if err != nil {
			t.Errorf("Failed opening repository: %s", err)
			return
		}

		volume, err := r.FindVolume(vol.ID)
		if err != nil {
			t.Errorf("Failed finding volume: %s", err)
			return
		}
		if volume.Name != vol.Name {
			t.Errorf("Failed verifying volume name: %s != %s", vol.Name, volume.Name)
		}
		if volume.Description != vol.Description {
			t.Errorf("Failed verifying volume description: %s != %s", vol.Description, volume.Description)
		}

		if r.Close() != nil {
			t.Errorf("Failed closing repository: %s", err)
		}
	}
}

func TestVolumeFind(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, _ := NewRepository(dir, testPassword)

	_, err = r.FindVolume("invalidID")
	if err != ErrVolumeNotFound {
		t.Errorf("Expected %v, got %v", ErrVolumeNotFound, err)
	}

	vol, _ := NewVolume("test", "")
	_ = r.AddVolume(vol)

	v, err := r.FindVolume("latest")
	if err != nil || v == nil {
		t.Errorf("Failed finding latest volume: %s %s", err, vol.ID)
	}

	if r.Close() != nil {
		t.Errorf("Failed closing repository: %s", err)
	}
}

func TestSnapshotRemove(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, _ := NewRepository(dir, testPassword)
	vol, _ := NewVolume("test", "")
	_ = r.AddVolume(vol)

	snapshot, _ := NewSnapshot("test_snapshot")
	_ = snapshot.Save(&r)
	_ = vol.AddSnapshot(snapshot.ID)

	snapshot2, _ := NewSnapshot("test_snapshot_too")
	_ = snapshot2.Save(&r)
	_ = vol.AddSnapshot(snapshot2.ID)

	err = vol.RemoveSnapshot(snapshot.ID)
	if err != nil {
		t.Errorf("Failed removing snapshot: %s", err)
	}

	_, _, err = r.FindSnapshot(snapshot2.ID)
	if err != nil {
		t.Errorf("Failed finding snapshot: %s", err)
	}

	err = vol.RemoveSnapshot("invalidID")
	if err == nil {
		t.Errorf("Expected no error, got: %s", err)
	}

	if r.Close() != nil {
		t.Errorf("Failed closing repository: %s", err)
	}
}
