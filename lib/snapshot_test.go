/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/minio/highwayhash"
)

func hashFile(path string) (string, error) {
	var key [32]byte
	hasher, err := highwayhash.New(key[:])
	if err != nil {
		return "", err
	}

	s, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func TestSnapshotCreate(t *testing.T) {
	testPassword := "this_is_a_password"

	tests := []struct {
		compression uint16
		parityParts uint
	}{
		{CompressionNone, 0},
		{CompressionGZip, 0},
		{CompressionLZMA, 0},
		{CompressionNone, 1},
		{CompressionGZip, 1},
	}
	for _, tt := range tests {
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
			progress := snapshot.Add(wd, []string{"snapshot_test.go", "snapshot.go"}, []string{}, r, &index, tt.compression, true, 1, tt.parityParts)
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

			for i, archive := range snapshotOriginal.Archives {
				if archive.Path != snapshot.Archives[i].Path {
					t.Errorf("Failed verifying snapshot archive: %s != %s", archive.Path, snapshot.Archives[i].Path)
					return
				}
				if archive.Size != snapshot.Archives[i].Size {
					t.Errorf("Failed verifying snapshot archive size: %d != %d", archive.Size, snapshot.Archives[i].Size)
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
				hash1, err := hashFile(file1)
				if err != nil {
					t.Errorf("Failed generating shasum for %s: %s", file1, err)
					return
				}
				hash2, err := hashFile(snapshotOriginal.Archives[i].Path)
				if err != nil {
					t.Errorf("Failed generating shasum for %s: %s", snapshotOriginal.Archives[i].Path, err)
					return
				}
				if hash1 != hash2 {
					t.Errorf("Failed verifying shasum: %s != %s", hash1, hash2)
					return
				}
			}
		}
	}
}

func TestSnapshotClone(t *testing.T) {
	snapshot, _ := NewSnapshot("test_snapshot")
	s, err := snapshot.Clone()
	if err != nil || s == nil {
		t.Errorf("Failed cloning snapshot: %s", err)
	}

	if snapshot.ID == s.ID {
		t.Errorf("ID conflict after cloning, duplicate snapshot ID %s", snapshot.ID)
	}
	if snapshot.Description != s.Description {
		t.Errorf("Description mismatch, got %s expected %s", s.Description, snapshot.Description)
	}

	for i, archive := range snapshot.Archives {
		if archive.Path != s.Archives[i].Path {
			t.Errorf("Failed verifying snapshot archive: %s != %s", archive.Path, s.Archives[i].Path)
			return
		}
		if archive.Size != snapshot.Archives[i].Size {
			t.Errorf("Failed verifying snapshot archive size: %d != %d", archive.Size, s.Archives[i].Size)
			return
		}
	}
}

func TestSnapshotFind(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, _ := NewRepository(dir, testPassword)
	vol, _ := NewVolume("test", "")
	r.AddVolume(vol)

	_, _, err = r.FindSnapshot("invalidID")
	if err != ErrSnapshotNotFound {
		t.Errorf("Expected %v, got %v", ErrSnapshotNotFound, err)
	}

	snapshot, _ := NewSnapshot("test_snapshot")
	snapshot.Save(&r)
	vol.AddSnapshot(snapshot.ID)

	_, s, err := r.FindSnapshot("latest")
	if err != nil || s == nil {
		t.Errorf("Failed finding latest snapshot: %s %s", err, snapshot.ID)
	}
}
