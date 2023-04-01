/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"os"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/action"
	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"

	"github.com/spf13/cobra"
)

var (
	catCmd = &cobra.Command{
		Use:   "cat [snapshot] [file]",
		Short: "print file",
		Long:  `The cat command prints a file on the standard output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("cat needs a snapshot ID and filename")
			}
			return executeCat(args[0], args[1])
		},
	}
)

func init() {
	RootCmd.AddCommand(catCmd)

	carapace.Gen(catCmd).PositionalCompletion(
		action.ActionSnapshots(catCmd, ""),
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			return action.ActionSnapshotPaths(catCmd, c.Args[0]).Invoke(c).ToMultiPartsA("/").StyleF(style.ForPathExt)
		}),
	)
}

func executeCat(snapshotID string, file string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	_, snapshot, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}

	if archive, ok := snapshot.Archives[file]; ok {
		b, _, err := knoxite.DecodeArchiveData(repository, *archive)
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(b)
		return err
	}

	return fmt.Errorf("%s: No such file or directory", file)
}
