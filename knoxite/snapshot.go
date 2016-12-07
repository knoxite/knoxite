/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package main

import (
	"fmt"

	"github.com/muesli/gotable"

	"github.com/knoxite/knoxite"
)

// CmdSnapshot describes the command
type CmdSnapshot struct {
	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("snapshot",
		"manage snapshots",
		"The snapshot command manages snapshots",
		&CmdSnapshot{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdSnapshot) Usage() string {
	return "[list|remove] VOLUME-ID"
}

// Execute this command
func (cmd CmdSnapshot) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return ErrMissingRepoLocation
	}

	switch args[0] {
	case "list":
		return cmd.list(args[1])
	case "remove":
		return cmd.remove(args[1])
	default:
		return fmt.Errorf(TUnknownCommand, cmd.Usage())
	}
}

func (cmd CmdSnapshot) remove(snapshotID string) error {
	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
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

func (cmd CmdSnapshot) list(volID string) error {
	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
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
	tab.Print()
	return nil
}
