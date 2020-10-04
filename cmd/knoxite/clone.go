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

			configureStoreOpts(cmd, &cloneOpts)
			return executeClone(args[0], args[1:], cloneOpts)
		},
	}
)

func init() {
	initStoreFlags(cloneCmd.Flags, &cloneOpts)
	RootCmd.AddCommand(cloneCmd)
}

func executeClone(snapshotID string, args []string, opts StoreOptions) error {
	targets := []string{}
	logger.Info("Collecting targets")
	for _, target := range args {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// acquire a shutdown lock. we don't want these next calls to be interrupted
	logger.Info("Acquiring shutdown lock")
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	logger.Info("Acquired and locked shutdown lock")

	logger.Info("Opening repository")
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	logger.Info("Opened repository")

	logger.Infof("Finding snapshot %s", snapshotID)
	volume, s, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}
	logger.Infof("Found snapshot %s", s.Description)

	logger.Info("Cloning snapshot")
	snapshot, err := s.Clone()
	if err != nil {
		return err
	}
	logger.Infof("Cloned snapshot. New snapshot: ID: %s, "+
		"Description: %s.", snapshot.ID, snapshot.Description)

	logger.Info("Opening chunk index")
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}
	logger.Info("Opened chunk index")
	// release the shutdown lock
	lock()
	logger.Info("Released shutdown lock")

	logger.Infof("Storing cloned snapshot %s", snapshot.ID)
	err = store(&repository, &chunkIndex, snapshot, targets, opts)
	if err != nil {
		return err
	}
	logger.Infof("Stored clone %s of snapshot %s", snapshot.ID, s.ID)

	// acquire another shutdown lock. we don't want these next calls to be interrupted
	logger.Info("Acquiring shutdown lock")
	lock = shutdown.Lock()
	if lock == nil {
		return nil
	}
	logger.Info("Acquired and locked shutdown lock")

	defer lock()
	defer logger.Info("Shutdown lock released")

	logger.Infof("Saving snapshot %s", snapshot.ID)
	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	logger.Info("Saved snapshot")

	logger.Infof("Adding snapshot to volume %s", volume.ID)
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	logger.Info("Added snapshot to volume")

	logger.Info("Saving chunk index")
	err = chunkIndex.Save(&repository)
	if err != nil {
		return err
	}
	logger.Info("Saved chunk index")

	logger.Info("Saving repository")
	err = repository.Save()
	if err != nil {
		return err
	}
	logger.Info("Saved repository")
	return nil
}
