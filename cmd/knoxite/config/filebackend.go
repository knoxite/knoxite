/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
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
	config := &Config{
		backend: fs,
		url:     u,
	}

	path, err := url.QueryUnescape(u.Path)
	if err != nil {
		return config, err
	}
	config.url.Path = path

	if !exist(path) {
		return config, nil
	}

	content, err := ioutil.ReadFile(config.url.Path)
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(content, config)
	config.backend = fs
	config.url = u
	return config, err
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

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(*config); err != nil {
		return err
	}

	return ioutil.WriteFile(config.URL().Path, buf.Bytes(), 0600)
}
