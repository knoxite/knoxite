/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020,      Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/muesli/gotable"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
)

// VolumeInitOptions holds all the options that can be set for the 'volume init' command.
type VolumeInitOptions struct {
	Description string
}

var (
	volumeInitOpts = VolumeInitOptions{}

	volumeCmd = &cobra.Command{
		Use:   "volume",
		Short: "manage volumes",
		Long:  `The volume command manages volumes`,
		RunE:  nil,
	}
	volumeInitCmd = &cobra.Command{
		Use:   "init <name>",
		Short: "initialize a new volume",
		Long:  `The init command initializes a new volume`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("init needs a name for the new volume")
			}
			return executeVolumeInit(args[0], volumeInitOpts.Description)
		},
	}
	volumeRemoveCmd = &cobra.Command{
		Use:   "remove <volume>",
		Short: "remove a volume from a repository",
		Long:  `The remove command removes a volume from a repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("remove needs a volume to work on")
			}
			return executeVolumeRemove(args[0])
		},
	}
	volumeListCmd = &cobra.Command{
		Use:   "list",
		Short: "list all volumes inside a repository",
		Long:  `The list command lists all volumes stored in a repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeVolumeList()
		},
	}
)

func init() {
	volumeInitCmd.Flags().StringVarP(&volumeInitOpts.Description, "desc", "d", "", "a description or comment for this volume")

	volumeCmd.AddCommand(volumeInitCmd)
	volumeCmd.AddCommand(volumeRemoveCmd)
	volumeCmd.AddCommand(volumeListCmd)
	RootCmd.AddCommand(volumeCmd)
}

func executeVolumeInit(name, description string) error {
	// we don't want these next calls to be interrupted
	logger.Info("Acquiring shutdown lock")
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	logger.Info("Acquired and locked shutdown lock")

	defer lock()
	defer logger.Info("Shutdown lock released")

	logger.Info("Opening repository")
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	logger.Info("Opened repository")

	logger.Infof("Creating volume %s", description)
	vol, err := knoxite.NewVolume(name, description)
	if err != nil {
		return err
	}
	logger.Infof("Created volume %s", vol.ID)

	logger.Infof("Adding volume %s to repository", vol.ID)
	err = repository.AddVolume(vol)
	if err != nil {
		return fmt.Errorf("Creating volume %s failed: %v", name, err)
	}
	logger.Info("Added volume to repository")

	annotation := "Name: " + vol.Name
	if len(vol.Description) > 0 {
		annotation += ", Description: " + vol.Description
	}
	fmt.Printf("Volume %s (%s) created\n", vol.ID, annotation)

	logger.Info("Saving repository")
	err = repository.Save()
	if err != nil {
		return err
	}
	logger.Info("Saved repository")
	return nil
}

func executeVolumeRemove(volumeID string) error {
	logger.Info("Opening repository")
	repo, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	logger.Info("Opened repository")

	logger.Info("Opening chunk index")
	chunkIndex, err := knoxite.OpenChunkIndex(&repo)
	if err != nil {
		return err
	}
	logger.Info("Opened chunk index")

	logger.Infof("Finding volume %s", volumeID)
	vol, err := repo.FindVolume(volumeID)
	if err != nil {
		return err
	}
	logger.Info("Found volume")

	logger.Infof("Iterating over all snapshots of volume %s to remove them", volumeID)
	for _, s := range vol.Snapshots {
		logger.Debugf("Removing snapshot %s", s)
		if err := vol.RemoveSnapshot(s); err != nil {
			return err
		}

		chunkIndex.RemoveSnapshot(s)
		logger.Debug("Removed snapshot")
	}
	logger.Info("Removed all snapshots from volume")

	logger.Infof("Removing volume %s from repository", volumeID)
	if err := repo.RemoveVolume(vol); err != nil {
		return err
	}
	logger.Info("Removed volume from repository")

	logger.Info("Saving chunk index")
	if err := chunkIndex.Save(&repo); err != nil {
		return err
	}
	logger.Info("Saved chunk index")

	logger.Info("Saving repository")
	if err := repo.Save(); err != nil {
		return err
	}
	logger.Info("Saved repository")

	fmt.Printf("Volume %s '%s' successfully removed\n", vol.ID, vol.Name)
	fmt.Println("Do not forget to run 'repo pack' to delete un-referenced chunks and free up storage space!")
	return nil
}

func executeVolumeList() error {
	logger.Info("Opening repository")
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	logger.Info("Opened repository")

	logger.Debug("Initializing new gotable for output")
	tab := gotable.NewTable([]string{"ID", "Name", "Description"},
		[]int64{-8, -32, -48}, "No volumes found. This repository is empty.")

	logger.Debug("Iterating over volumes to print details")
	for _, volume := range repository.Volumes {
		tab.AppendRow([]interface{}{volume.ID, volume.Name, volume.Description})
	}

	logger.Debug("Printing volume list output")
	err = tab.Print()
	if err != nil {
		return err
	}
	logger.Debug("Printed output")
	return nil
}
