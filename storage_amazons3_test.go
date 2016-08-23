/*
 * knoxite
 *     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"net/url"
	"testing"
)

func TestNewStorageAmazonS3(t *testing.T) {
	// run setup_minio_test_environment.sh to start minio with these keys
	accesskey := "USWUXHGYZQYFYFFIT3RE"
	secretkey := "MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03"

	validURL, err := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = NewStorageAmazonS3(*validURL)
	if err != nil {
		t.Error(err)
	}

	missingUsername, err := url.Parse("s3://" + secretkey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = NewStorageAmazonS3(*missingUsername)
	if err == nil {
		t.Error(ErrInvalidRepositoryURL)
	}

	missingPassword, err := url.Parse("s3://" + accesskey + "@127.0.0.1:9000/us-east-1/test")
	if err != nil {
		t.Error(err)
	}
	_, err = NewStorageAmazonS3(*missingPassword)
	if err == nil {
		t.Error(ErrInvalidRepositoryURL)
	}

	missingRegion, err := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/test")
	if err != nil {
		t.Error(err)
	}
	_, err = NewStorageAmazonS3(*missingRegion)
	if err == nil {
		t.Error(ErrInvalidRepositoryURL)
	}

	missingBucket, err := url.Parse("s3://" + accesskey + ":" + secretkey + "@127.0.0.1:9000/us-east-1/")
	if err == nil {
		t.Error(ErrInvalidRepositoryURL)
	}
	_, err = NewStorageAmazonS3(*missingBucket)
	if err != nil {
		t.Error(err)
	}
}
