/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package main

import (
	"errors"
	"fmt"

	"github.com/muesli/goprogressbar"

	"github.com/knoxite/knoxite"
)

// Error declarations
var (
	ErrTargetMissing = errors.New("please specify a directory to restore to (--target)")
)

// CmdRestore describes the command
type CmdRestore struct {
	Target string `short:"t" long:"target" description:"Directory to restore to"`

	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("restore",
		"restore a snapshot",
		"The restore command restores a snapshot to a directory",
		&CmdRestore{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdRestore) Usage() string {
	return "SNAPSHOT-ID"
}

// Execute this command
func (cmd CmdRestore) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return ErrMissingRepoLocation
	}
	if cmd.Target == "" {
		return ErrTargetMissing
	}

	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err == nil {
		_, snapshot, ferr := repository.FindSnapshot(args[0])
		if ferr != nil {
			return ferr
		}

		progress, derr := knoxite.DecodeSnapshot(repository, *snapshot, cmd.Target)
		if derr != nil {
			return derr
		}
		pb := goprogressbar.NewProgressBar("", 0, 0, 60)
		stats := knoxite.Stats{}
		lastPath := ""

		for p := range progress {
			pb.Total = int64(p.Size)
			pb.Current = int64(p.Transferred)
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
			if p.Size == p.Transferred {
				// We have just finished restoring an item
				stats.Add(p.Statistics)
			}

			pb.Print()
		}
		fmt.Println()
		fmt.Println("Restore done:", stats.String())
		return nil
	}

	return err
}
