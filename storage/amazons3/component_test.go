// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/knoxite/knoxite/storage"
)

var (
	backendTest *storage.BackendTest
)

func TestMain(m *testing.M) {
	amazons3url := os.Getenv("KNOXITE_AMAZONS3NG_URL")
	if len(amazons3url) == 0 {
		panic("no backend configured")
	}

	parsedUrl, err := url.Parse(amazons3url)
	if err != nil {
		panic("invalid url")
	}
	parsedUrl.Path = path.Join(parsedUrl.Path, storage.RandomSuffix())
	amazons3url = parsedUrl.String()

	backendTest = &storage.BackendTest{
		URL:         amazons3url,
		Protocols:   []string{"amazons3"},
		Description: "Amazon S3 Storage Backend (using AWS SDK)",
		TearDown:    func(tb *storage.BackendTest) {},
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
