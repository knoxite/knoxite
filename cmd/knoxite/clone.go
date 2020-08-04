/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"path/filepath"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
)

var (
	cloneOpts = StoreOptions{}

	cloneCmd = &cobra.Command{
		Use:   "clone <snapshot> <dir/file> [...]",
		Short: "clone a snapshot",
		Long:  `The clone command clones an existing snapshot and adds a file or directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("clone needs to know which snapshot to clone")
			}
			if len(args) < 2 {
				return fmt.Errorf("clone needs to know which files and/or directories to work on")
			}
			cloneOpts = configureStoreOpts(cmd, cloneOpts)

			return executeClone(args[0], args[1:], cloneOpts)
		},
	}
)

func init() {
	initStoreFlags(cloneCmd.Flags)
	RootCmd.AddCommand(cloneCmd)
}

func executeClone(snapshotID string, args []string, opts StoreOptions) error {
	targets := []string{}
	for _, target := range args {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	volume, s, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}
	snapshot, err := s.Clone()
	if err != nil {
		return err
	}
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}
	// release the shutdown lock
	lock()

	err = store(&repository, &chunkIndex, snapshot, targets, opts)
	if err != nil {
		return err
	}

	// acquire another shutdown lock. we don't want these next calls to be interrupted
	lock = shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	err = chunkIndex.Save(&repository)
	if err != nil {
		return err
	}
	return repository.Save()
}
