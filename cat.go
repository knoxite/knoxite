/*
 * knoxite
 *     Copyright (c) 2016-2018, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"os"

	"github.com/knoxite/knoxite/lib"

	"github.com/spf13/cobra"
)

var (
	catCmd = &cobra.Command{
		Use:   "cat <snapshot> <file>",
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
}

func executeCat(snapshotID string, file string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	_, snapshot, ferr := repository.FindSnapshot(snapshotID)
	if ferr != nil {
		return ferr
	}

	if archive, ok := snapshot.Archives[file]; ok {
		b, _, erra := knoxite.DecodeArchiveData(repository, *archive)
		if erra != nil {
			return erra
		}

		_, err = os.Stdout.Write(b)
		return err
	}

	return fmt.Errorf("%s: No such file or directory", file)
}
