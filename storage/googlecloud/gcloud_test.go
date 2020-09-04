// +build backend

/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package googlecloud

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	googlestorage "cloud.google.com/go/storage"
	"github.com/knoxite/knoxite/storage"
	"google.golang.org/api/iterator"
)

var (
	backendTest *storage.BackendTest
)

func TestMain(m *testing.M) {

	// create the json key file needed for Google Cloud Storage authentication
	jsonKeyTemplate := `
	{
		"type": "service_account",
		"project_id": "%%project_id%%",
		"private_key_id": "%%private_key_id%%",
		"private_key": "%%private_key%%",
		"client_email": "%%client_email%%",
		"client_id": "%%client_id%%",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "%%client_x509_cert_url%%"
	  }
	`

	replacer := strings.NewReplacer("%%project_id%%", os.Getenv("KNOXITE_GC_KEY_PROJECT_ID"),
		"%%private_key_id%%", os.Getenv("KNOXITE_GC_KEY_PRIVATE_KEY_ID"),
		"%%private_key%%", os.Getenv("KNOXITE_GC_KEY_PRIVATE_KEY"),
		"%%client_email%%", os.Getenv("KNOXITE_GC_KEY_CLIENT_EMAIL"),
		"%%client_id%%", os.Getenv("KNOXITE_GC_KEY_CLIENT_ID"),
		"%%client_x509_cert_url%%", os.Getenv("KNOXITE_GC_KEY_CLIENT_X509_CERT_URL"))

	tmpFile, err := ioutil.TempFile(os.TempDir(), "googlecloud-json-key-*.json")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())

	jsonKey := []byte(replacer.Replace(jsonKeyTemplate))
	if _, err = tmpFile.Write(jsonKey); err != nil {
		panic(err)
	}

	// create a random folder name to avoid collisions
	rnd := storage.RandomSuffix()

	googlecloudurl := os.Getenv("KNOXITE_GOOGLECLOUD_URL")
	if len(googlecloudurl) == 0 {
		panic("no backend configured")
	}

	googlecloudurl = strings.Replace(googlecloudurl, "%%json-key%%", url.QueryEscape(tmpFile.Name()), 1)

	if err := tmpFile.Close(); err != nil {
		panic(err)
	}

	backendTest = &storage.BackendTest{
		URL:         googlecloudurl + rnd + "/",
		Protocols:   []string{"googlecloudstorage"},
		Description: "Google Cloud Storage",
		TearDown: func(tb *storage.BackendTest) {
			db := tb.Backend.(*GoogleCloudStorage)
			it := db.bucket.Objects(context.Background(), &googlestorage.Query{Prefix: rnd})
			for {
				objAttrs, err := it.Next()
				if err != nil && err != iterator.Done {
					panic(err)
				}
				if err == iterator.Done {
					break
				}
				if err := db.bucket.Object(objAttrs.Name).Delete(context.Background()); err != nil {
					panic(err)
				}
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
