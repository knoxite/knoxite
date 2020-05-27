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
	"github.com/spf13/cobra"
)

var (
	verifyCmd = &cobra.Command{
		Use:   "verify [<volume> [<snapshot>]]",
		Short: "verify a repo, volume or snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return executeVerifyRepo()
			} else if len(args) == 1 {
				return executeVerifyVolume(args[0])
			} else if len(args) == 2 {
				return executeVerifySnapshot(args[0], args[1])
			}
			return nil
		},
	}
)

func init() {
	RootCmd.AddCommand(verifyCmd)
}

func executeVerifyRepo() error {

	return nil
}

func executeVerifyVolume(volumeId string) error {

	return nil
}

func executeVerifySnapshot(volumeId string, snapshotId string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)

	if err == nil {
		_, snapshot, ferr := repository.FindSnapshot(snapshotId)
		if ferr != nil {
			return ferr
		}

		for _, value := range snapshot.Archives {
			fmt.Println(knoxite.VerifyArchive(repository, *value))
		}

	}

	return nil
}
