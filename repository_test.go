/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRepositoryCreate(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, err := NewRepository(dir, testPassword)
	if err != nil {
		t.Errorf("Failed creating repository: %s", err)
		return
	}
	if r.Close() != nil {
		t.Errorf("Failed closing repository: %s", err)
	}

	r, err = OpenRepository(dir, testPassword, false)
	if err != nil {
		t.Errorf("Failed opening repository: %s", err)
		return
	}
	if r.Close() != nil {
		t.Errorf("Failed closing repository: %s", err)
	}
}

func TestRepositoryCreateError(t *testing.T) {
	testPassword := "this_is_a_password"

	_, err := NewRepository("invalidprotocol://foo", testPassword)
	if err != ErrInvalidRepositoryURL {
		t.Errorf("Expected %v, got %v", ErrInvalidRepositoryURL, err)
	}
}

func TestRepositoryIsEmpty(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	r, _ := NewRepository(dir, testPassword)
	if !r.IsEmpty() {
		t.Error("Repository should be empty")
	}

	vol, _ := NewVolume("test", "")
	_ = r.AddVolume(vol)
	if !r.IsEmpty() {
		t.Error("Repository should be empty")
	}

	snapshot, _ := NewSnapshot("test_snapshot")
	_ = snapshot.Save(&r)
	_ = vol.AddSnapshot(snapshot.ID)
	if r.IsEmpty() {
		t.Error("Repository should not be empty")
	}

	if r.Close() != nil {
		t.Errorf("Failed closing repository: %s", err)
	}
}

func TestRepositoryChangePassword(t *testing.T) {
	testPassword := "this_is_a_password"
	newPassword := "this_is_another_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	repo, err := NewRepository(dir, testPassword)
	if err != nil {
		t.Errorf("Failed creating repository: %s", err)
		return
	}
	err = repo.Close()
	if err != nil {
		t.Errorf("Failed closing repository: %s", err)
		return
	}

	repo, err = OpenRepository(dir, testPassword, true)
	if err != nil {
		t.Errorf("Failed opening repository: %s", err)
		return
	}

	if repo.ChangePassword(newPassword) != nil {
		t.Errorf("Failed to change repository password: %s", err)
		return
	}
	repo.Close()

	_, err = OpenRepository(dir, testPassword, false)
	if err == nil {
		t.Errorf("Repository can still be opened with the old password after changing it: %s", err)
		return
	}

	repo, err = OpenRepository(dir, newPassword, false)
	if err != nil {
		t.Errorf("Failed opening repository with new password after changing it: %s", err)
		return
	}
	repo.Close()
}

func TestRepositoryLock(t *testing.T) {
	testPassword := "this_is_a_password"

	dir, err := ioutil.TempDir("", "knoxite")
	if err != nil {
		t.Errorf("Failed creating temporary dir for repository: %s", err)
		return
	}
	defer os.RemoveAll(dir)

	// create a new repo, which will acquire a lock on it
	repo, err := NewRepository(dir, testPassword)
	if err != nil {
		t.Errorf("Failed creating repository: %s", err)
		return
	}

	// try opening the repo r/w while it's locked
	_, err = OpenRepository(dir, testPassword, true)
	if err == nil {
		t.Error("Repository should have been locked!")
		return
	}

	// but r/o access should still work
	_, err = OpenRepository(dir, testPassword, false)
	if err != nil {
		t.Errorf("Failed opening repository: %s", err)
		return
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Failed closing repository: %s", err)
		return
	}

	// try opening the repo after the lock has been released
	repo, err = OpenRepository(dir, testPassword, true)
	if err != nil {
		t.Errorf("Failed opening repository: %s", err)
		return
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Failed closing repository: %s", err)
		return
	}
}
