package main

import (
	"errors"
	"fmt"

	"github.com/knoxite/knoxite"
	"github.com/muesli/gotable"
)

// CmdSnapshot describes the command
type CmdSnapshot struct {
	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("snapshot",
		"show snapshots",
		"The snapshots command lists all snapshots stored in a repository",
		&CmdSnapshot{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdSnapshot) Usage() string {
	return "[list] VOLUME-ID"
}

// Execute this command
func (cmd CmdSnapshot) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}

	switch args[0] {
	case "list":
		return cmd.list(args[1])
	default:
		return fmt.Errorf(TUnknownCommand, cmd.Usage())
	}
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
