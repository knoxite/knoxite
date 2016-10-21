/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package main

import (
	"fmt"
	"path/filepath"
)

// CmdClone describes the command
type CmdClone struct {
	store *CmdStore

	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("clone",
		"clone a snapshot",
		"The clone command clones an existing snapshot and adds a file or directory",
		&CmdClone{store: &CmdStore{}, global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdClone) Usage() string {
	return "SNAPSHOT-ID DIR/FILE [DIR/FILE] [...]"
}

// Execute this command
func (cmd CmdClone) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return ErrMissingRepoLocation
	}

	targets := []string{}
	for _, target := range args[1:] {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// filter here? exclude/include?

	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return err
	}
	volume, s, err := repository.FindSnapshot(args[0])
	if err != nil {
		return err
	}
	snapshot, err := s.Clone()
	if err != nil {
		return err
	}
	err = cmd.store.store(&repository, snapshot, targets)
	if err != nil {
		return err
	}
	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	return repository.Save()
}
