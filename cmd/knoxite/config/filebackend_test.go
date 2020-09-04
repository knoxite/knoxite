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

func TestFileLoad(t *testing.T) {
	u := &url.URL{
		Scheme: "file",
		Path:   url.QueryEscape("foobar"),
	}
	backend := NewFileBackend()

	_, err := backend.Load(u)
	if err != nil {
		t.Error("Loading an non-existing config file should not return an error")
	}

	// try to load the config from a relative path
	u.Path = url.QueryEscape(filepath.Join("testdata", "knoxite.conf"))
	backend = NewFileBackend()
	conf, err := backend.Load(u)
	if err != nil {
		t.Errorf("Error loading config file fixture from relative path %s. %v", u, err)
	}
	repo, ok := conf.Repositories["knoxitetest"]
	if !ok {
		t.Errorf("There should exist an repoconfig aliased with knox")
	}
	if repo.Url != "/tmp/knoxitetest" {
		t.Errorf("Expected '/tmp/koxite/ as repo url, got: %s", repo.Url)
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
	if !repo.Pedantic {
		t.Errorf("Expected 'pedantic' to be true, got: %v", repo.Pedantic)
	}

	// try to load the config from an absolute path using a URI
	cwd, _ := os.Getwd()
	u.Path = url.QueryEscape(filepath.Join(cwd, "testdata", "knoxite.conf"))

	backend = NewFileBackend()
	conf, err = backend.Load(u)
	if err != nil {
		t.Errorf("Error loading config file fixture from absolute path %s. %v", u, err)
	}
	repo, ok = conf.Repositories["knoxitetest"]
	if !ok {
		t.Errorf("There should exist an repoconfig aliased with knox")
	}
	if repo.Url != "/tmp/knoxitetest" {
		t.Errorf("Expected '/tmp/koxite/ as repo url, got: %s", repo.Url)
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
}

func TestFileSave(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "knoxitetest")
	if err != nil {
		t.Error("Could not create temp directory")
	}

	u := &url.URL{
		Scheme: "file",
		Path:   url.QueryEscape(filepath.Join("testdata", "knoxite.conf")),
	}

	backend := NewFileBackend()
	c, err := backend.Load(u)
	if err != nil {
		t.Errorf("Failed to load config fixture from relative path %s: %v", u, err)
	}

	// Save the config file to a new absolute path using a URL
	p := filepath.Join(tmpdir, "knoxite.conf")
	u.Path = url.QueryEscape(p)

	backend = NewFileBackend()
	err = backend.Save(c)
	if err != nil {
		t.Errorf("Failed to save the config to %s", u)
	}
	if !exist(p) {
		t.Errorf("Configuration file wasn't saved to %s", p)
	}
	c, err = backend.Load(u)
	if err != nil {
		t.Errorf("Failed to load config fixture from absolute path %s: %v", u, err)
	}

	// Save the config file to a new absolute path using a regular path
	u.Scheme = ""

	err = backend.Save(c)
	if err != nil {
		t.Errorf("Failed to save the config to %s", p)
	}
	if !exist(p) {
		t.Errorf("Configuration file wasn't saved to %s", p)
	}
}
