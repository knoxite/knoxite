/*
 * knoxite
 *     Copyright (c) 2017-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestChunkIndexReindex(t *testing.T) {
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
	index, _ := OpenChunkIndex(&r)
	wd, _ := os.Getwd()

	opts := StoreOptions{
		CWD:         wd,
		Paths:       []string{"snapshot_test.go", "snapshot.go"},
		Excludes:    []string{},
		Compress:    CompressionNone,
		Encrypt:     EncryptionAES,
		DataParts:   1,
		ParityParts: 0,
	}

	progress := snapshot.Add(r, &index, opts)
	for p := range progress {
		if p.Error != nil {
			t.Errorf("Failed adding to snapshot: %s", p.Error)
		}
	}

	_ = snapshot.Save(&r)
	_ = vol.AddSnapshot(snapshot.ID)
	_ = index.Save(&r)
	_ = r.Save()

	newindex, err := OpenChunkIndex(&r)
	if err != nil {
		t.Errorf("Failed reopening chunk-index: %s", err)
	}
	err = newindex.reindex(&r)
	if err != nil {
		t.Errorf("Failed reindexing chunk-index: %s", err)
	}
}

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
	_ = r.AddVolume(vol)

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

	opts := StoreOptions{
		CWD:         wd,
		Paths:       []string{"snapshot_test.go", "snapshot.go"},
		Excludes:    []string{},
		Compress:    CompressionNone,
		Encrypt:     EncryptionAES,
		DataParts:   1,
		ParityParts: 0,
	}

	progress := snapshot.Add(r, &index, opts)
	for p := range progress {
		if p.Error != nil {
			t.Errorf("Failed adding to snapshot: %s", p.Error)
		}
	}

	_ = snapshot.Save(&r)
	_ = vol.AddSnapshot(snapshot.ID)

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
