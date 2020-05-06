// +build backend

/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package azure

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/Azure/azure-storage-file-go/azfile"

	"github.com/knoxite/knoxite/storage"
)

var (
	backendTest *storage.BackendTest
)

func TestMain(m *testing.M) {
	// create a random folder to avoid collisions
	rnd := storage.RandomSuffix()

	azurefileurl := os.Getenv("KNOXITE_AZURE_FILE_URL")
	if len(azurefileurl) == 0 {
		panic("no backend configured")
	}

	backendTest = &storage.BackendTest{
		URL:         azurefileurl + rnd,
		Protocols:   []string{"azurefile"},
		Description: "Azure file storage",
		TearDown: func(tb *storage.BackendTest) {
			err := DeleteFolder(tb, rnd)
			if err != nil {
				panic(err)
			}
		},
	}

	storage.RunBackendTester(backendTest, m)
}

func DeleteFolder(tb *storage.BackendTest, p string) error {
	db := tb.Backend.(*AzureFileStorage)
	pipeline := azfile.NewPipeline(&db.credential, azfile.PipelineOptions{})

	shareUrl := azfile.NewShareURL(db.endpoint, pipeline)
	directoryUrl := shareUrl.NewDirectoryURL(p)

	content, err := directoryUrl.ListFilesAndDirectoriesSegment(
		context.Background(),
		azfile.Marker{},
		azfile.ListFilesAndDirectoriesOptions{})

	if err != nil {
		panic(err)
	}

	for _, file := range content.FileItems {
		u := db.endpoint
		u.Path = path.Join(u.Path, p+"/"+file.Name)

		_, err := azfile.NewFileURL(u, pipeline).Delete(context.Background())

		if err != nil {
			return err
		}
	}
	for _, dir := range content.DirectoryItems {
		err = DeleteFolder(tb, p+"/"+dir.Name)
		if err != nil {
			return err
		}
	}

	return directoryUrl.Delete(context.Background())
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
