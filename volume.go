/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"

	"github.com/klauspost/shutdown2"
	"github.com/muesli/gotable"
	"github.com/spf13/cobra"

	knoxite "github.com/knoxite/knoxite/lib"
)

// VolumeInitOptions holds all the options that can be set for the 'volume init' command
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
	volumeCmd.AddCommand(volumeListCmd)
	RootCmd.AddCommand(volumeCmd)
}

func executeVolumeInit(name, description string) error {
	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		vol, verr := knoxite.NewVolume(name, description)
		if verr == nil {
			verr = repository.AddVolume(vol)
			if verr != nil {
				return fmt.Errorf("Creating volume %s failed: %v", name, verr)
			}

			annotation := "Name: " + vol.Name
			if len(vol.Description) > 0 {
				annotation += ", Description: " + vol.Description
			}
			fmt.Printf("Volume %s (%s) created\n", vol.ID, annotation)
			return repository.Save()
		}
	}
	return err
}

func executeVolumeList() error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	tab := gotable.NewTable([]string{"ID", "Name", "Description"},
		[]int64{-8, -32, -48}, "No volumes found. This repository is empty.")
	for _, volume := range repository.Volumes {
		tab.AppendRow([]interface{}{volume.ID, volume.Name, volume.Description})
	}

	tab.Print()
	return nil
}
