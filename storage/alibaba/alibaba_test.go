// +build backend

/*
 * knoxite
 *     Copyright (c) 2021, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021, Nicolas Martin <penguwin@penguwin.eu>
 *     // TODO
 *
 *   For license see LICENSE
 */

package alibaba

import (
	"os"
	"testing"

	"github.com/knoxite/knoxite/storage"
)

var (
	backendTest *storage.BackendTest
)

func TestMain(m *testing.M) {
	// create a random path suffix to avoid collisions
	rnd := storage.RandomSuffix()

	alibabaurl := os.Getenv("KNOXITE_ALIBABA_URL")
	if len(alibabaurl) == 0 {
		panic("no backend configured")
	}

	backendTest = &storage.BackendTest{
		URL:         alibaba + rnd,
		Protocols:   []string{"alibaba"},
		Description: "Alibaba Storage",
		TearDown: func(tb *storage.BackendTest) {
			// TODO:
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
