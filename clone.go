/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"path/filepath"

	"github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"

	knoxite "github.com/knoxite/knoxite/lib"
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
			return executeClone(args[0], args[1:], cloneOpts)
		},
	}
)

func init() {
	cloneCmd.Flags().StringVarP(&storeOpts.Description, "desc", "d", "", "a description or comment for this volume")
	cloneCmd.Flags().StringVarP(&storeOpts.Compression, "compression", "c", "", "compression algo to use: none (default), gzip")
	cloneCmd.Flags().StringVarP(&storeOpts.Encryption, "encryption", "e", "", "encryption algo to use: aes (default), none")
	cloneCmd.Flags().UintVarP(&storeOpts.FailureTolerance, "tolerance", "t", 0, "failure tolerance against n backend failures")

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

	// filter here? exclude/include?

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
