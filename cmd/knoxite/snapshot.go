/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"

	"github.com/muesli/gotable"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
)

var (
	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "manage snapshots",
		Long:  `The snapshot command manages snapshots`,
		RunE:  nil,
	}
	snapshotListCmd = &cobra.Command{
		Use:   "list <volume>",
		Short: "list all snapshots inside a volume",
		Long:  `The list command lists all snapshots stored in a volume`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("list needs a volume ID to work on")
			}
			return executeSnapshotList(args[0])
		},
	}
	snapshotRemoveCmd = &cobra.Command{
		Use:   "remove <snapshot>",
		Short: "remove a snapshot",
		Long:  `The remove command deletes a snapshot from a volume`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("remove needs a snapshot ID to work on")
			}
			return executeSnapshotRemove(args[0])
		},
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRemoveCmd)
	RootCmd.AddCommand(snapshotCmd)
}

func executeSnapshotRemove(snapshotID string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}

	volume, snapshot, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}

	err = volume.RemoveSnapshot(snapshot.ID)
	if err != nil {
		return err
	}

	chunkIndex.RemoveSnapshot(snapshot.ID)
	err = chunkIndex.Save(&repository)
	if err != nil {
		return err
	}

	err = repository.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Snapshot %s removed: %s\n", snapshot.ID, snapshot.Stats.String())
	fmt.Println("Do not forget to run 'repo pack' to delete un-referenced chunks and free up storage space!")
	return nil
}

func executeSnapshotList(volID string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	volume, err := repository.FindVolume(volID)
	if err != nil {
		return err
	}

	tab := gotable.NewTable([]string{"ID", "Date", "Original Size", "Storage Size", "Description"},
		[]int64{-8, -19, 13, 12, -48}, "No snapshots found. This volume is empty.")
	totalSize := uint64(0)
	totalStorageSize := uint64(0)

	for _, snapshotID := range volume.Snapshots {
		snapshot, err := volume.LoadSnapshot(snapshotID, &repository)
		if err != nil {
			return err
		}
		tab.AppendRow([]interface{}{
			snapshot.ID,
			snapshot.Date.Format(timeFormat),
			knoxite.SizeToString(snapshot.Stats.Size),
			knoxite.SizeToString(snapshot.Stats.StorageSize),
			snapshot.Description})
		totalSize += snapshot.Stats.Size
		totalStorageSize += snapshot.Stats.StorageSize
	}

	tab.SetSummary([]interface{}{"", "", knoxite.SizeToString(totalSize), knoxite.SizeToString(totalStorageSize), ""})
	_ = tab.Print()
	return nil
}
