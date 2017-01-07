/*
 * knoxite
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestChunkIndexPack(t *testing.T) {
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

	snapshot, _ := NewSnapshot("test_snapshot")
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
	progress := snapshot.Add(wd, []string{"snapshot_test.go", "snapshot.go"}, r, &index, false, true, 1, 0)
	for p := range progress {
		if p.Error != nil {
			t.Errorf("Failed adding to snapshot: %s", p.Error)
		}
	}

	snapshot.Save(&r)
	vol.AddSnapshot(snapshot.ID)

	err = vol.RemoveSnapshot(snapshot.ID)
	if err != nil {
		t.Errorf("Failed removing snapshot: %s", err)
	}
	index.RemoveSnapshot(snapshot.ID)

	_, err = index.Pack(&r)
	if err != nil {
		t.Errorf("Packing chunk index failed: %s", err)
	}
}
