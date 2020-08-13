/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const testPassword = "test"

func TestAESBackendLoad(t *testing.T) {
	cwd, _ := os.Getwd()

	// try to load the config from an absolute path using a URI
	u := &url.URL{
		Scheme: "crypto",
		User:   url.User(url.QueryEscape(testPassword)),
		// Using QueryEscape here as PathEscape won't escape some symbols like ':'.
		Path: url.PathEscape(filepath.Join(cwd, "testdata", "knoxite-crypto.conf")),
	}

	backend, _ := NewAESBackend(u)
	conf, err := backend.Load(u)
	if err != nil {
		t.Errorf("Error loading config file fixture from absolute path %s: %v", u.String(), err)
	}

	repo, ok := conf.Repositories["knoxitetest"]
	if !ok {
		t.Errorf("There should exist an repoconfig aliased with 'knoxitetest'")
	}
	if repo.Url != "/tmp/knoxitetest" {
		t.Errorf("Expected '/tmp/koxitetest as repo url, got: %s", repo.Url)
	}
	if repo.Compression != "gzip" {
		t.Errorf("Expected gzip as compression type, got: %s", repo.Compression)
	}
	if repo.Encryption != "aes" {
		t.Errorf("Expected aes as encryption type, got: %s", repo.Encryption)
	}
	if repo.Tolerance != 0 {
		t.Errorf("Expected repoTolerance of 0, got: %v", repo.Tolerance)
	}
	excludes := []string{"just", "an", "example"}
	if !reflect.DeepEqual(repo.StoreExcludes, excludes) {
		t.Errorf("Store Excludes did not match:\nExpected: %v\nGot: %v", excludes, repo.StoreExcludes)
	}
	if !reflect.DeepEqual(repo.RestoreExcludes, excludes) {
		t.Errorf("Restore Excludes did not match:\nExpected: %v\nGot: %v", excludes, repo.RestoreExcludes)
	}

	// try to load the config with a wrong password
	u = &url.URL{
		Scheme: "crypto",
		User:   url.User(url.QueryEscape("wrongPassword")),
		Path:   url.QueryEscape(filepath.Join(cwd, "testdata", "knoxite-crypto.conf")),
	}

	backend, _ = NewAESBackend(u)
	_, err = backend.Load(u)
	if err == nil || err.Error() != "cipher: message authentication failed" {
		t.Errorf("loading the config file with an invalid password should fail. %v", err)
	}
}

func TestAESBackendSave(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "knoxitetest")
	if err != nil {
		t.Error("Could not create temp directory")
	}

	cwd, _ := os.Getwd()
	u := &url.URL{
		Scheme: "crypto",
		User:   url.User(url.QueryEscape(testPassword)),
		Path:   url.QueryEscape(filepath.Join(cwd, "testdata", "knoxite-crypto.conf")),
	}

	backend, _ := NewAESBackend(u)
	c, err := backend.Load(u)
	if err != nil {
		t.Errorf("Failed to load config fixture from relative path %s: %v", u, err)
	}

	// Save the config file to a new absolute path using a URL
	absPath := filepath.Join(tmpdir, "knoxite-crypto.conf")
	u.Path = url.QueryEscape(absPath)

	backend, _ = NewAESBackend(u)
	err = backend.Save(c)
	if err != nil {
		t.Errorf("failed to save the config to %s", u)
	}
	if !exist(absPath) {
		t.Errorf("configuration file wasn't saved to %s", absPath)
	}

	ok, err := IsEncrypted(u)
	if !ok {
		t.Errorf("encrypted config header not added. %v", err)
	}
}
