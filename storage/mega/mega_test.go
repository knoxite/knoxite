// +build backend

/*
 * knoxite
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */
package mega

import (
	"os"
	"testing"

	"github.com/knoxite/knoxite/storage"
)

func TestMain(m *testing.M) {
	// create a random bucket name to avoid collisions
	rnd := storage.RandomSuffix()

	megaurl := os.Getenv("KNOXITE_MEGA_URL")
	if len(megaurl) == 0 {
		panic("no backend configured")
	}

	backendTest = &storage.BackendTest{
		URL:         megaurl + rnd,
		Protocols:   []string{"mega"},
		Description: "mega.nz storage",
		TearDown: func(tb *storage.BackendTest) {
			db := tb.Backend.(*MegaStorage)
			err := db.DeleteFile(db.Path)
			if err != nil {
				panic(err)
			}
		},
	}

	storage.RunBackendTester(backendTest, m)
}

var (
	backendTest *storage.BackendTest
)

func TestStorageNewBackend(t *testing.T) {
	backendTest.NewBackendTest(t)
}

func TestStorageLocation(t *testing.T) {
	backendTest.LocationTest(t)
}

func TestStorageProtocols(t *testing.T) {
	backendTest.ProtocolsTest(t)
}

func TestStorageDescription(t *testing.T) {
	backendTest.DescriptionTest(t)
}

func TestStorageInitRepository(t *testing.T) {
	backendTest.InitRepositoryTest(t)
}

func TestStorageSaveRepository(t *testing.T) {
	backendTest.SaveRepositoryTest(t)
}

func TestStorageAvailableSpace(t *testing.T) {
	backendTest.AvailableSpaceTest(t)
}

func TestStorageSaveSnapshot(t *testing.T) {
	backendTest.SaveSnapshotTest(t)
}

func TestStorageStoreChunk(t *testing.T) {
	backendTest.StoreChunkTest(t)
}

func TestStorageDeleteChunk(t *testing.T) {
	backendTest.DeleteChunkTest(t)
}

func TestStorageLocking(t *testing.T) {
	backendTest.LockingTest(t)
}
