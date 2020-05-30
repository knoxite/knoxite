/*
* knoxite
*     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
*
*   For license see LICENSE
 */
package cfg

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	conf, err := New("/foobar")
	if err != nil {
		panic(err)
	}

	if conf.Type() != FileConf {
		t.Error("Backend for '/foobar' should be a FileBackend")
	}

	conf, err = New("file:///foobar")
	if err != nil {
		panic(err)
	}

	if conf.Type() != FileConf {
		t.Error("Backend for 'file:///foobar' should be a FileBackend")
	}

	cwd, _ := os.Getwd()
	p := filepath.Join(cwd, "testdata/knoxite-crypto.conf")
	conf, err = New(url.QueryEscape(p))
	if conf.Type() != CryptoConf {
		t.Errorf("Backend for '%s' should be an AESBackend", p)
	}

	conf, err = New("mem:")
	if err != nil {
		panic(err)
	}

	if conf.Type() != MemoryConf {
		t.Error("Backend for 'mem:' should be a MemoryBackend")
	}

	conf, err = New("c:\\foobar")
	if err == nil {
		t.Error("Not a valid URL, should return an error")
	}

	conf, err = New("")
	if err == nil {
		t.Error("Not a valid URL, should return an error")
	}
}

func TestLoad(t *testing.T) {
	conf, err := New("mem://")
	if err != nil {
		panic(err)
	}
	err = conf.Load()
	if err != nil {
		t.Error("Failed loading the configuration")
	}
	if conf.URL().Scheme != "mem" {
		t.Error("Config URL didn't change")
	}
}
