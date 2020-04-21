// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package backblaze

import (
	"os"
	"testing"

	"github.com/knoxite/knoxite/storage"
)

var (
	backendTest *storage.BackendTest
)

func TestMain(m *testing.M) {
	// create a random bucket name to avoid collisions
	rnd := storage.RandomSuffix()

	backblazeurl := os.Getenv("KNOXITE_BACKBLAZE_URL")
	if len(backblazeurl) == 0 {
		panic("no backend configured")
	}

	backendTest = &storage.BackendTest{
		URL:         backblazeurl + rnd,
		Protocols:   []string{"backblaze"},
		Description: "Backblaze Storage",
		TearDown: func(tb *storage.BackendTest) {
			db := tb.Backend.(*BackblazeStorage)
			list, err := db.Bucket.ListFileNames("", 128)
			if err != nil {
				panic(err)
			}

			for _, v := range list.Files {
				ids, lerr := db.Bucket.ListFileVersions(v.Name, "", 128)
				if lerr != nil {
					panic(lerr)
				}
				for _, id := range ids.Files {
					db.Bucket.DeleteFileVersion(v.Name, id.ID)
				}
			}

			err = db.Bucket.Delete()
			if err != nil {
				panic(err)
			}
		},
	}

	storage.RunBackendTester(backendTest, m)
}

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

func TestAvailableSpace(t *testing.T) {
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
