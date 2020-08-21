/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

// FileBackend implements a filesystem backend for the configuration.
type FileBackend struct{}

// NewFileBackend returns a FileBackend that handles loading and
// saving files from the local filesytem.
func NewFileBackend() *FileBackend {
	return &FileBackend{}
}

func (fs *FileBackend) Type() int {
	return FileConf
}

// Load a config from a URL.
func (fs *FileBackend) Load(u *url.URL) (*Config, error) {
	var config Config

	path, err := url.QueryUnescape(u.Path)
	if err != nil {
		return nil, err
	}
	u.Path = path

	if !exist(u.Path) {
		return &Config{url: u}, nil
	}

	j, err := ioutil.ReadFile(u.Path)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(j, &config)
	if err != nil {
		return nil, err
	}
	config.backend = fs
	config.url = u
	return &config, nil
}

// Save config.
func (fs *FileBackend) Save(config *Config) error {
	path, err := url.QueryUnescape(config.URL().Path)
	if err != nil {
		return err
	}
	config.URL().Path = path

	cfgDir := filepath.Dir(config.URL().Path)
	if !exist(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return err
		}
	}

	j, err := json.MarshalIndent(config, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(config.URL().Path, j, 0644)
	}

	return err
}
