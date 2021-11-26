/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/knoxite/knoxite/cmd/knoxite/utils"
	gap "github.com/muesli/go-app-paths"
)

const appName = "knoxite"

// Available configuration backends.
const (
	FileConf = iota
	CryptoConf
	MemoryConf
)

var cfgFileName = "knoxite.conf"

// The RepoConfig struct contains all the default values for a a repository.
type RepoConfig struct {
	Url             string   `toml:"url" comment:"Repository directory to backup to/restore from"`
	Compression     string   `toml:"compression" comment:"Compression algo to use: none (default), flate, gzip, lzma, zlib, zstd"`
	Tolerance       uint     `toml:"tolerance" comment:"Failure tolerance against n backend failures"`
	Encryption      string   `toml:"encryption" comment:"Encryption algo to use: aes (default), none"`
	Pedantic        bool     `toml:"pedantic" comment:"Stop backup operation after the first error occurred"`
	VerifyStore     bool     `toml:"verify_store" comment:"Verify each written chunk for the store operation"`
	StoreExcludes   []string `toml:"store_excludes" comment:"Specify excludes for the store operation"`
	RestoreExcludes []string `toml:"restore_excludes" comment:"Specify excludes for the restore operation"`
}

type Config struct {
	Repositories map[string]RepoConfig `toml:"repositories"`
	backend      ConfigBackend
	url          *url.URL
}

// ConfigBackend is the interface implemented by the configuration backends.
//
// Backends are responsible for loading and saving the Config struct to
// the local filesystem, the network, etc.
type ConfigBackend interface {
	Load(*url.URL) (*Config, error)
	Save(*Config) error
	Type() int
}

func (c *Config) Save() error {
	return c.backend.Save(c)
}

// Load the configuration.
//
// The backend loaded will be responsible for loading it
// from the given URL.
func (c *Config) Load() error {
	config, err := c.backend.Load(c.url)
	if err != nil {
		return err
	}
	c.Repositories = config.Repositories
	return nil
}

// Type of the backend which is currently being used.
func (c *Config) Type() int {
	return c.backend.Type()
}

// SetURL updates the configuration URL.
//
// Next time the config is loaded or saved
// the new URL will be used.
func (c *Config) SetURL(u string) error {
	url, err := utils.PathToUrl(u)
	if err != nil {
		return err
	}
	c.url = url
	return nil
}

// URL currently being used.
func (c *Config) URL() *url.URL {
	return c.url
}

// New returns a new Config struct
//
// The URL will be matched against all the supported backends and the first
// backend that can handle the URL scheme will be loaded.
func New(url string) (*Config, error) {
	config := &Config{}
	var backend ConfigBackend

	if url == "" {
		return nil, fmt.Errorf("empty URL provided but not supported")
	}

	err := config.SetURL(url)
	if err != nil {
		return nil, err
	}

	switch config.url.Scheme {
	case "", "file":
		if ok, _ := IsEncrypted(config.url); ok {
			// fmt.Println("Loading encrypted configuration file")
			if backend, err = NewAESBackend(config.url); err != nil {
				return nil, fmt.Errorf("error loading the AES configuration backend: %v", err)
			}
		} else {
			backend = NewFileBackend()
		}
	case "mem":
		backend = NewMemBackend()
	case "crypto":
		if backend, err = NewAESBackend(config.url); err != nil {
			return nil, fmt.Errorf("error loading the AES configuration backend: %v", err)
		}
	default:
		return nil, fmt.Errorf("configuration backend '%s' not supported", config.url.Scheme)
	}

	config.backend = backend
	return config, nil
}

// DefaultPath returns Knoxite's default config path.
//
// The path returned is OS dependant. If there's an error
// while trying to figure out the OS dependant path, "knoxite.conf"
// in the current working dir is returned.
func DefaultPath() string {
	userScope := gap.NewScope(gap.User, appName)
	path, err := userScope.ConfigPath(cfgFileName)
	if err != nil {
		return cfgFileName
	}

	return path
}

// Lookup tries to find the config file.
//
// If a config file is found in the current working directory, that's returned.
// Otherwise we try to locate it following an OS dependant:
//
// Unix:
//   - ~/.config/knoxite/knoxite.conf
// macOS:
//   - ~/Library/Preferences/knoxite/knoxite.conf
// Windows:
//   - %LOCALAPPDATA%/knoxite/Config/knoxite.conf
//
// If no valid config file is found, an empty string is returned.
func Lookup() string {
	paths := []string{}
	defaultPath := DefaultPath()
	if exist(defaultPath) {
		paths = append(paths, defaultPath)
	}

	// Prepend ./knoxite.conf to the search path if exists, takes priority
	// over the rest.
	cwd, err := os.Getwd()
	if err != nil {
		// log.Printf("Error getting current working directory: %v", err)
		cwd = "."
	}
	cwdCfg := filepath.Join(cwd, cfgFileName)
	if exist(cwdCfg) {
		paths = append([]string{cwdCfg}, paths...)
	}
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func exist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
