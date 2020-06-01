/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"

	"github.com/knoxite/knoxite"
	"github.com/muesli/goprogressbar"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type VerifyOptions struct {
	Percentage int
}

var (
	verifyOpts = VerifyOptions{}

	verifyCmd = &cobra.Command{
		Use:   "verify [<volume> [<snapshot>]]",
		Short: "verify a repo, volume or snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return executeVerifyRepo()
			} else if len(args) == 1 {
				return executeVerifyVolume(args[0])
			} else if len(args) == 2 {
				return executeVerifySnapshot(args[0], args[1], verifyOpts)
			}
			return nil
		},
	}
)

func initVerifyFlags(f func() *pflag.FlagSet) {
	f().IntVar(&verifyOpts.Percentage, "percentage", 70, "How many archives to be checked between 0 and 100")
}

func init() {
	initVerifyFlags(verifyCmd.Flags)
	RootCmd.AddCommand(verifyCmd)
}

func executeVerifyRepo() error {

	return nil
}

func executeVerifyVolume(volumeId string) error {

	return nil
}

func executeVerifySnapshot(volumeId string, snapshotId string, opts VerifyOptions) error {
	errors := make([]error, 0)
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		progress, err := knoxite.VerifySnapshot(repository, snapshotId, opts.Percentage)
		if err != nil {
			errors = append(errors, err)
			return err
		}

		pb := &goprogressbar.ProgressBar{Total: 1000, Width: 40}
		lastPath := ""
	

		for p := range progress {
			if p.Error != nil {
				fmt.Println()
				errors = append(errors, p.Error)
			}

			pb.Total = int64(p.CurrentItemStats.Size)
			pb.Current = int64(p.CurrentItemStats.Transferred)
			pb.PrependText = fmt.Sprintf("%s / %s",
				knoxite.SizeToString(uint64(pb.Current)),
				knoxite.SizeToString(uint64(pb.Total)))

			if p.Path != lastPath {
				// We have just started restoring a new item
				if len(lastPath) > 0 {
					fmt.Println()
				}
				lastPath = p.Path
				pb.Text = p.Path
			}

			pb.LazyPrint()

		}
		fmt.Println()
		fmt.Printf("Verify done: %d errors\n", len(errors))
		return nil
	}
	return err
}
