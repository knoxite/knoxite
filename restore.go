/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"errors"
	"fmt"

	"github.com/muesli/goprogressbar"
	"github.com/spf13/cobra"

	knoxite "github.com/knoxite/knoxite/lib"
)

// Error declarations
var (
	ErrTargetMissing = errors.New("please specify a directory to restore to")
)

var (
	restoreCmd = &cobra.Command{
		Use:   "restore <snapshot> <destination>",
		Short: "restore a snapshot",
		Long:  `The restore command restores a snapshot to a directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("restore needs to know which snapshot to work on")
			}
			if len(args) < 2 {
				return ErrTargetMissing
			}
			return executeRestore(args[0], args[1])
		},
	}
)

func init() {
	RootCmd.AddCommand(restoreCmd)
}

func executeRestore(snapshotID, target string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		_, snapshot, ferr := repository.FindSnapshot(snapshotID)
		if ferr != nil {
			return ferr
		}

		progress, derr := knoxite.DecodeSnapshot(repository, snapshot, target)
		if derr != nil {
			return derr
		}
		pb := goprogressbar.NewProgressBar("", 0, 0, 60)
		stats := knoxite.Stats{}
		lastPath := ""

		for p := range progress {
			pb.Total = int64(p.CurrentItemStats.Size)
			pb.Current = int64(p.CurrentItemStats.Transferred)
			pb.RightAlignedText = fmt.Sprintf("%s / %s  %s/s",
				knoxite.SizeToString(uint64(pb.Current)),
				knoxite.SizeToString(uint64(pb.Total)),
				knoxite.SizeToString(p.TransferSpeed()))

			if p.Path != lastPath {
				// We have just started restoring a new item
				if len(lastPath) > 0 {
					fmt.Println()
				}
				lastPath = p.Path
				pb.Text = p.Path
			}
			if p.CurrentItemStats.Size == p.CurrentItemStats.Transferred {
				// We have just finished restoring an item
				stats.Add(p.TotalStatistics)
			}

			pb.Print()
		}
		fmt.Println()
		fmt.Println("Restore done:", stats.String())
		return nil
	}

	return err
}
