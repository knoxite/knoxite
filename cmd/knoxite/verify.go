/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/knoxite/knoxite"
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
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)

	if err == nil {
		_, snapshot, ferr := repository.FindSnapshot(snapshotId)
		if ferr != nil {
			return ferr
		}

		// get all keys of the snapshot Archives
		archives := make([]string, 0, len(snapshot.Archives))
		for key := range snapshot.Archives {
			archives = append(archives, key)
		}

		if opts.Percentage > 100 {
			opts.Percentage = 100
		} else if opts.Percentage < 0 {
			opts.Percentage = 0
		}

		// select len(keys)*percentage unique keys to verify
		nrOfSelectedArchives := int(math.Ceil(float64(len(archives)*opts.Percentage) / 100.0))

		// we use a map[string]bool as a Set implementation
		selectedArchives := make(map[string]bool)
		for len(selectedArchives) < nrOfSelectedArchives {
			idx := rand.Intn(len(archives))
			selectedArchives[archives[idx]] = true
		}

		errors := make([]struct {
			error
			string
		}, 0)

		mutex := sync.Mutex{}

		for archiveKey := range selectedArchives {
			err := knoxite.VerifyArchive(repository, *snapshot.Archives[archiveKey])
			if err != nil {
				mutex.Lock()
				errors = append(errors, struct {
					error
					string
				}{err, archiveKey})
			}
		}

		if len(errors) == 0 {
			fmt.Println("No errors found")
		} else {
			fmt.Printf("Verification failed! Errors in the following files:\n")
			for _, err := range errors {
				fmt.Printf("    %s: %s\n", err.string, err.error)
			}
		}

	}

	return nil
}
