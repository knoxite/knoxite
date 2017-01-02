/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func shasumFile(path string) (string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func TestCreateSnapshot(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	snapshotOriginal := &Snapshot{}
	{
		r, err := NewRepository(dir, testPassword)
		if err != nil {
			t.Errorf("Failed creating repository: %s", err)
			return
		}
		vol, err := NewVolume("test_name", "test_description")
		if err != nil {
			t.Errorf("Failed creating volume: %s", err)
			return
		}
		err = r.AddVolume(vol)
		if err != nil {
			t.Errorf("Failed creating volume: %s", err)
			return
		}
		snapshot, err := NewSnapshot("test_snapshot")
		if err != nil {
			t.Errorf("Failed creating snapshot: %s", err)
			return
		}
		index, err := OpenChunkIndex(&r)
		if err != nil {
			t.Errorf("Failed opening chunk-index: %s", err)
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			t.Errorf("Failed getting working dir: %s", err)
			return
		}
		progress := snapshot.Add(wd, []string{"snapshot_test.go"}, r, &index, false, true, 1, 0)
		for p := range progress {
			if p.Error != nil {
				t.Errorf("Failed adding to snapshot: %s", p.Error)
			}
		}

		err = snapshot.Save(&r)
		if err != nil {
			t.Errorf("Failed saving snapshot: %s", err)
		}
		err = vol.AddSnapshot(snapshot.ID)
		if err != nil {
			t.Errorf("Failed adding snapshot to volume: %s", err)
		}
		err = r.Save()
		if err != nil {
			t.Errorf("Failed saving volume: %s", err)
			return
		}
		err = index.Save(&r)
		if err != nil {
			t.Errorf("Failed saving chunk-index: %s", err)
			return
		}

		snapshotOriginal = snapshot
	}

	{
		r, err := OpenRepository(dir, testPassword)
		if err != nil {
			t.Errorf("Failed opening repository: %s", err)
			return
		}

		_, snapshot, err := r.FindSnapshot(snapshotOriginal.ID)
		if err != nil {
			t.Errorf("Failed finding snapshot: %s", err)
			return
		}
		if !snapshot.Date.Equal(snapshotOriginal.Date) {
			t.Errorf("Failed verifying snapshot date: %v != %v", snapshot.Date, snapshotOriginal.Date)
		}
		if snapshot.Description != snapshotOriginal.Description {
			t.Errorf("Failed verifying snapshot description: %s != %s", snapshot.Description, snapshotOriginal.Description)
		}

		for i, archive := range snapshot.Archives {
			if archive.Path != snapshotOriginal.Archives[i].Path {
				t.Errorf("Failed verifying snapshot archive: %s != %s", archive.Path, snapshotOriginal.Archives[i].Path)
				return
			}
			if archive.Size != snapshotOriginal.Archives[i].Size {
				t.Errorf("Failed verifying snapshot archive size: %d != %d", archive.Size, snapshotOriginal.Archives[i].Size)
				return
			}
		}

		targetdir, err := ioutil.TempDir("", "knoxite.target")
		if err != nil {
			t.Errorf("Failed creating temporary dir for restore: %s", err)
			return
		}
		defer os.RemoveAll(targetdir)

		progress, err := DecodeSnapshot(r, snapshot, targetdir)
		if err != nil {
			t.Errorf("Failed restoring snapshot: %s", err)
			return
		}
		for range progress {
		}

		for i, archive := range snapshot.Archives {
			file1 := filepath.Join(targetdir, archive.Path)
			sha1, err := shasumFile(file1)
			if err != nil {
				t.Errorf("Failed generating shasum for %s: %s", file1, err)
				return
			}
			sha2, err := shasumFile(snapshotOriginal.Archives[i].Path)
			if err != nil {
				t.Errorf("Failed generating shasum for %s: %s", snapshotOriginal.Archives[i].Path, err)
				return
			}
			if sha1 != sha2 {
				t.Errorf("Failed verifying shasum: %s != %s", sha1, sha2)
				return
			}
		}
	}
}

func TestFindUnknownSnapshot(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, err := NewRepository(dir, testPassword)
	if err != nil {
		t.Errorf("Failed creating repository: %s", err)
		return
	}

	vol, err := NewVolume("test", "")
	if err != nil {
		t.Errorf("Failed creating volume: %s", err)
		return
	}
	r.AddVolume(vol)

	_, _, err = r.FindSnapshot("invalidID")
	if err != ErrSnapshotNotFound {
		t.Errorf("Expected %v, got %v", ErrSnapshotNotFound, err)
	}
}
