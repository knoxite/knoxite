/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *                   2020, Sergio Rubio <sergio@rubio.im>
 *
 *   For license see LICENSE
 */
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	conf, err := New("/foobar")
	if err != nil {
		t.Fatalf("Failed creating file backend: %s", err)
	}

	if conf.Type() != FileConf {
		t.Error("Backend for '/foobar' should be a FileBackend")
	}

	conf, err = New("file:///foobar")
	if err != nil {
		t.Fatalf("Failed creating file backend: %s", err)
	}

	if conf.Type() != FileConf {
		t.Error("Backend for 'file:///foobar' should be a FileBackend")
	}

	cwd, _ := os.Getwd()
	p := filepath.Join(cwd, "testdata/knoxite-crypto.conf")
	conf, err = New(p)
	if err != nil {
		t.Fatalf("Failed loading crypto config backend: %s", err)
	}
	if conf.Type() != CryptoConf {
		t.Errorf("Backend for '%s' should be an AESBackend", p)
	}

	conf, err = New("mem://")
	if err != nil {
		t.Fatalf("Failed creating mem backend: %s", err)
	}

	if conf.Type() != MemoryConf {
		t.Error("Backend for 'mem:' should be a MemoryBackend")
	}

	_, err = New("~/foobar")
	if err != nil {
		t.Fatalf("Failed here: %v", err)
	}

	_, err = New("c:\\foobar")
	if err != nil {
		t.Errorf("Failed to create backend for valid windows url: %s", err)
	}

	_, err = New("..\\Foobar")
	if err != nil {
		t.Errorf("Failed to create backend for valid windows url: %s", err)
	}

	_, err = New("\\Foobar")
	if err != nil {
		t.Errorf("Failed to create backend for valid windows url: %s", err)
	}

	_, err = New("")
	if err == nil {
		t.Error("Not a valid URL, should return an error")
	}
}

func TestLoad(t *testing.T) {
	conf, err := New("mem://")
	if err != nil {
		t.Fatalf("Failed creating mem backend: %s", err)
	}
	err = conf.Load()
	if err != nil {
		t.Error("Failed loading the configuration")
	}
	if conf.URL().Scheme != "mem" {
		t.Error("Config URL didn't change")
	}
}
