// +build ci

/*
 * knoxite
 *     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
 *
 *   For license see LICENSE.txt
 */

package amazons3

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	mrand "math/rand"
	"net/url"
	"testing"

	knoxite "github.com/knoxite/knoxite/lib"
)

func TestNewStorageAmazonS3(t *testing.T) {
	// run setup_minio_test_environment.sh to start minio with these keys
	accesskey := "USWUXHGYZQYFYFFIT3RE"
	secretkey := "MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03"

	validURL, err := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = (&StorageAmazonS3{}).NewBackend(*validURL)
	if err != nil {
		t.Error(err)
	}

	missingUsername, err := url.Parse("s3://" + secretkey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = (&StorageAmazonS3{}).NewBackend(*missingUsername)
	if err == nil {
		t.Error(knoxite.ErrInvalidRepositoryURL)
	}

	missingPassword, err := url.Parse("s3://" + accesskey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = (&StorageAmazonS3{}).NewBackend(*missingPassword)
	if err == nil {
		t.Error(knoxite.ErrInvalidRepositoryURL)
	}

	missingRegion, err := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/test")
	if err != nil {
		t.Error(err)
	}
	_, err = (&StorageAmazonS3{}).NewBackend(*missingRegion)
	if err == nil {
		t.Error(knoxite.ErrInvalidRepositoryURL)
	}
}

func TestStorageAmazonS3Location(t *testing.T) {
	s3url := createValidStorageURL()

	s3, _ := (&StorageAmazonS3{}).NewBackend(*s3url)
	if s3.Location() != s3url.String() {
		t.Errorf("Expected %v, got %v", s3.Location(), s3url.String())
	}
}

func TestStorageAmazonS3Close(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()

	if s3.Close() != nil {
		t.Error("There should not be an error on Close")
	}
}

func TestStorageAmazonS3Protocols(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()

	if len(s3.Protocols()) != 2 {
		t.Error("Invalid amount of protocols")
	}

	if s3.Protocols()[0] != "s3" && s3.Protocols()[1] != "s3s" {
		t.Error("Invalid protocols")
	}
}

func TestStorageAmazonS3Description(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()

	if s3.Description() != "Amazon S3 Storage" {
		t.Error("Invalid Description")
	}
}

func TestStorageAmazonS3InitRepository(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}
}

func TestStorageAmazonS3SaveRepository(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnd := make([]byte, 256)
	rand.Read(rnd)

	err := s3.SaveRepository(rnd)
	if err != nil {
		t.Error(err)
	}

	data, err := s3.LoadRepository()
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnd) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnd[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3LoadRepository(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnd := make([]byte, 256)
	rand.Read(rnd)

	err := s3.SaveRepository(rnd)
	if err != nil {
		t.Error(err)
	}

	data, err := s3.LoadRepository()
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnd) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnd[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3SaveSnapshot(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	rndid := make([]byte, 8)
	rand.Read(rndid)
	id := hex.EncodeToString(rndid)

	err := s3.SaveSnapshot(id, rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := s3.LoadSnapshot(id)
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnddata) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3LoadSnapshot(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	rndid := make([]byte, 8)
	rand.Read(rndid)
	id := hex.EncodeToString(rndid)

	err := s3.SaveSnapshot(id, rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := s3.LoadSnapshot(id)
	if err != nil {
		t.Error(err)
	}

	if len(data) != len(rnddata) {
		t.Error("Data length missmatch")
	}

	for i := 0; i != len(data); i++ {
		if data[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3StoreChunk(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	size, err := s3.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}
	if size != uint64(len(rnddata)) {
		t.Error("Data length missmatch")
	}

	// Test to store the same chunk twice. Size should be 0
	size, err = s3.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}
	if size != 0 {
		t.Error("Already exisiting chunks should not be overwritten")
	}

	data, err := s3.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}
	if len(*data) != len(rnddata) {
		t.Error("Data length missmatch")
	}
	for i := 0; i != len(*data); i++ {
		if (*data)[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3LoadChunk(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	_, err := s3.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}

	data, err := s3.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}
	if len(*data) != len(rnddata) {
		t.Error("Data length missmatch")
	}
	for i := 0; i != len(*data); i++ {
		if (*data)[i] != rnddata[i] {
			t.Error("Data missmatch")
		}
	}
}

func TestStorageAmazonS3DeleteChunk(t *testing.T) {
	s3 := createValidStorageAmazonS3Object()
	if err := s3.InitRepository(); err != nil {
		t.Error(err)
	}

	rnddata := make([]byte, 256)
	rand.Read(rnddata)

	totalParts := uint(mrand.Int())
	// get a random part number which is smaller than the totalParts number
	part := uint(mrand.Intn(int(totalParts)))

	shasumdata := sha256.Sum256(rnddata)
	shasum := hex.EncodeToString(shasumdata[:])

	_, err := s3.StoreChunk(shasum, part, totalParts, &rnddata)
	if err != nil {
		t.Error(err)
	}

	_, err := s3.LoadChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}

	err = s3.DeleteChunk(shasum, part, totalParts)
	if err != nil {
		t.Error(err)
	}

	_, err := s3.LoadChunk(shasum, part, totalParts)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func createValidStorageAmazonS3Object() knoxite.Backend {
	s3, _ := (&StorageAmazonS3{}).NewBackend(*createValidStorageURL())
	return s3
}

func createValidStorageURL() *url.URL {
	accesskey := "USWUXHGYZQYFYFFIT3RE"
	secretkey := "MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03"

	// create a random bucket name every time to avoid collisions
	rnd := make([]byte, 8)
	rand.Read(rnd)
	bucket := hex.EncodeToString(rnd)

	validURL, _ := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/us-east-1/" + bucket)
	return validURL
}
