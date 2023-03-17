/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/muesli/gotable"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/action"
	"github.com/knoxite/knoxite/cmd/knoxite/utils"
)

var (
	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "manage snapshots",
		Long:  `The snapshot command manages snapshots`,
		RunE:  nil,
	}
	snapshotListCmd = &cobra.Command{
		Use:   "list [volume]",
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
		Use:   "remove [snapshot]",
		Short: "remove a snapshot",
		Long:  `The remove command deletes a snapshot from a volume`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("remove needs a snapshot ID to work on")
			}
			return executeSnapshotRemove(args[0])
		},
	}
	snapshotDiffCmd = &cobra.Command{
		Use:   "diff [snapshot1] [snapshot2]",
		Short: "shows difference between to snapshots",
		Long:  `The diff command shows the difference between two specified snapshots`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("diff needs two snapshot IDs to work on")
			}
			return executeSnapshotDiff(args[0], args[1])
		},
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRemoveCmd)
	snapshotCmd.AddCommand(snapshotDiffCmd)
	RootCmd.AddCommand(snapshotCmd)

	carapace.Gen(snapshotListCmd).PositionalCompletion(
		action.ActionVolumes(snapshotListCmd),
	)

	carapace.Gen(snapshotRemoveCmd).PositionalCompletion(
		action.ActionSnapshots(snapshotRemoveCmd, ""),
	)
}

func executeSnapshotDiff(snapshotID1, snapshotID2 string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}

	if snapshotID1 == snapshotID2 {
		return fmt.Errorf("diff needs to distinct snapshot IDs")
	}

	tab := gotable.NewTable([]string{"Snapshot (" + snapshotID1 + ")", "Snapshot (" + snapshotID2 + ")", "Operation", "User", "Group", "Perms"},
		[]int64{-48, -48, -9, -8, -5, -10},
		"No files found.")

	_, snapshot1, err := repository.FindSnapshot(snapshotID1)
	if err != nil {
		return err
	}

	_, snapshot2, err := repository.FindSnapshot(snapshotID2)
	if err != nil {
		return err
	}

	if snapshot1.Date.After(snapshot2.Date) {
		snapshot1, snapshot2 = snapshot2, snapshot1
	}

	paths1 := make([]string, 0, len(snapshot1.Archives))
	for k := range snapshot1.Archives {
		paths1 = append(paths1, k)
	}

	paths2 := make([]string, 0, len(snapshot2.Archives))
	for k := range snapshot2.Archives {
		paths2 = append(paths2, k)
	}

	for path, archive := range snapshot2.Archives {
		username := strconv.FormatInt(int64(archive.UID), 10)
		u, err := user.LookupId(username)
		if err == nil {
			username = u.Username
		}
		groupname := strconv.FormatInt(int64(archive.GID), 10)
		operation := "Unchanged"

		if utils.Contains(paths1, path) {
			if archive.ModTime > snapshot1.Archives[path].ModTime {
				operation = "Modified"
			} else {
				operation = "Unchanged"
			}

			tab.AppendRow([]interface{}{
				archive.Path,
				archive.Path,
				operation,
				username,
				groupname,
				archive.Mode,
			})
		} else {
			operation = "Created"

			tab.AppendRow([]interface{}{
				"-",
				archive.Path,
				operation,
				username,
				groupname,
				archive.Mode,
			})
		}

	}

	for path, archive := range snapshot1.Archives {
		username := strconv.FormatInt(int64(archive.UID), 10)
		u, err := user.LookupId(username)
		if err == nil {
			username = u.Username
		}
		groupname := strconv.FormatInt(int64(archive.GID), 10)

		if !utils.Contains(paths2, path) {
			tab.AppendRow([]interface{}{
				archive.Path,
				"-",
				"Deleted",
				username,
				groupname,
				archive.Mode,
			})
		}
	}

	_ = tab.Print()

	return nil
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
