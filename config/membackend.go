/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import "net/url"

// MemBackend implements a dummy memory backend for the configuration
type MemBackend struct {
	conf *Config
}

// NewMemBackend returns a backend that handles loading and saving
// the configuration from memory
func NewMemBackend() *MemBackend {
	return &MemBackend{conf: &Config{}}
}

func (m *MemBackend) Type() int {
	return MemoryConf
}

// Load the config from memory
//
// No need to do anything here, already loaded
func (m *MemBackend) Load(u *url.URL) (*Config, error) {
	return m.conf, nil
}

// Save the config to memory
//
// No need to do anything special here, already in memory
func (m *MemBackend) Save(config *Config) error {
	return nil
}
