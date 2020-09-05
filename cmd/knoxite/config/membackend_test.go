/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"net/url"
	"path/filepath"
	"testing"
)

func TestMemLoad(t *testing.T) {
	u := &url.URL{
		Scheme: "mem",
	}

	backend := NewMemBackend()
	_, err := backend.Load(u)
	if err != nil {
		t.Error("Loading an invalid config file should return an error")
	}
}

func TestMemSave(t *testing.T) {
	path := filepath.Join("testdata", "foobar")
	u := &url.URL{
		Path: url.QueryEscape(path),
	}

	backend := NewMemBackend()
	conf := &Config{url: u}

	if backend.Save(conf) != nil {
		t.Errorf("Failed to save the config to memory")
	}

	if exist(path) {
		t.Error("Configuration file should not exist")
	}
}
