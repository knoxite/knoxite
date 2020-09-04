/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */

package main

import (
	"errors"
	"fmt"

	"github.com/muesli/goprogressbar"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/knoxite/knoxite"
)

// Error declarations
var (
	ErrTargetMissing = errors.New("please specify a directory to restore to")
)

type RestoreOptions struct {
	Excludes []string
	Pedantic bool
}

var (
	restoreOpts = RestoreOptions{}

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

			configureRestoreOpts(cmd, &restoreOpts)
			return executeRestore(args[0], args[1], restoreOpts)
		},
	}
)

// configureRestoreOpts will compare the values from the configuration file and
// the user set command line flags.
// Values set via the command line flags will overwrite settings stored in the
// configuration file.
func configureRestoreOpts(cmd *cobra.Command, opts *RestoreOptions) {
	if rep, ok := cfg.Repositories[globalOpts.Alias]; ok {
		if !cmd.Flags().Changed("excludes") {
			opts.Excludes = rep.RestoreExcludes
		}
	}
}

func initRestoreFlags(f func() *pflag.FlagSet) {
	f().StringArrayVarP(&restoreOpts.Excludes, "excludes", "x", []string{}, "list of excludes")
	f().BoolVar(&restoreOpts.Pedantic, "pedantic", false, "exit on first error")
}

func init() {
	initRestoreFlags(restoreCmd.Flags)
	RootCmd.AddCommand(restoreCmd)
}

func executeRestore(snapshotID, target string, opts RestoreOptions) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		_, snapshot, err := repository.FindSnapshot(snapshotID)
		if err != nil {
			return err
		}

		progress, err := knoxite.DecodeSnapshot(repository, snapshot, target, opts.Excludes)
		if err != nil {
			return err
		}

		pb := &goprogressbar.ProgressBar{Total: 1000, Width: 40}
		stats := knoxite.Stats{}
		lastPath := ""

		errs := make(map[string]error)
		for p := range progress {
			if p.Error != nil {
				if restoreOpts.Pedantic {
					fmt.Println()
					return p.Error
				}
				errs[p.Path] = p.Error
				stats.Errors++
			}

			pb.Total = int64(p.CurrentItemStats.Size)
			pb.Current = int64(p.CurrentItemStats.Transferred)
			pb.PrependText = fmt.Sprintf("%s / %s  %s/s",
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

			pb.LazyPrint()
		}
		fmt.Println()
		fmt.Println("Restore done:", stats.String())
		for file, err := range errs {
			fmt.Printf("'%s' failed to restore: %v\n", file, err)
		}

		return nil
	}

	return err
}
