/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/utils"
)

// Error declarations
var (
	ErrRedundancyAmount = errors.New("failure tolerance can't be equal or higher as the number of storage backends")
)

// StoreOptions holds all the options that can be set for the 'store' command.
type StoreOptions struct {
	Description      string
	Compression      string
	Encryption       string
	FailureTolerance uint
	Excludes         []string
}

var (
	storeOpts = StoreOptions{}

	storeCmd = &cobra.Command{
		Use:   "store <volume> <dir/file> [...]",
		Short: "store files/directories",
		Long:  `The store command creates a snapshot of a file or directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("store needs to know which volume to create a snapshot in")
			}
			if len(args) < 2 {
				return fmt.Errorf("store needs to know which files and/or directories to work on")
			}

			configureStoreOpts(cmd, &storeOpts)
			return executeStore(args[0], args[1:], storeOpts)
		},
	}
)

// configureStoreOpts will compare the settings from the configuration file and
// the user set command line flags.
// Values set via the command line flags will overwrite settings stored in the
// configuration file.
func configureStoreOpts(cmd *cobra.Command, opts *StoreOptions) {
	if rep, ok := cfg.Repositories[globalOpts.Repo]; ok {
		if !cmd.Flags().Changed("compression") {
			opts.Compression = rep.Compression
		}
		if !cmd.Flags().Changed("encryption") {
			opts.Encryption = rep.Encryption
		}
		if !cmd.Flags().Changed("tolerance") {
			opts.FailureTolerance = rep.Tolerance
		}
		if !cmd.Flags().Changed("excludes") {
			opts.Excludes = rep.StoreExcludes
		}
	}
}

func initStoreFlags(f func() *pflag.FlagSet) {
	f().StringVarP(&storeOpts.Description, "desc", "d", "", "a description or comment for this volume")
	f().StringVarP(&storeOpts.Compression, "compression", "c", "", "compression algo to use: none (default), flate, gzip, lzma, zlib, zstd")
	f().StringVarP(&storeOpts.Encryption, "encryption", "e", "", "encryption algo to use: aes (default), none")
	f().UintVarP(&storeOpts.FailureTolerance, "tolerance", "t", 0, "failure tolerance against n backend failures")
	f().StringArrayVarP(&storeOpts.Excludes, "excludes", "x", []string{}, "list of excludes")
}

func init() {
	initStoreFlags(storeCmd.Flags)
	RootCmd.AddCommand(storeCmd)
}

func store(repository *knoxite.Repository, chunkIndex *knoxite.ChunkIndex, snapshot *knoxite.Snapshot, targets []string, opts StoreOptions) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()

	wd, gerr := os.Getwd()
	if gerr != nil {
		return gerr
	}

	if len(repository.BackendManager().Backends)-int(opts.FailureTolerance) <= 0 {
		return ErrRedundancyAmount
	}
	compression, err := utils.CompressionTypeFromString(opts.Compression)
	if err != nil {
		return err
	}
	encryption, err := utils.EncryptionTypeFromString(opts.Encryption)
	if err != nil {
		return err
	}

	tol := uint(len(repository.BackendManager().Backends) - int(opts.FailureTolerance))

	progress := snapshot.Add(wd, targets, opts.Excludes, *repository, chunkIndex,
		compression, encryption,
		tol, opts.FailureTolerance)

	consoleRenderer := ConsoleRenderer{}
	consoleRenderer.Init()

	output := knoxite.DefaultOutput{
		Renderers: knoxite.Renderers{&consoleRenderer},
	}

	err = output.Render(progress, cancel)
	fmt.Printf("\nSnapshot %s created: %s\n", snapshot.ID, snapshot.Stats.String())

	return err
}

func executeStore(volumeID string, args []string, opts StoreOptions) error {
	targets := []string{}
	for _, target := range args {
		if absTarget, err := filepath.Abs(target); err == nil {
			target = absTarget
		}
		targets = append(targets, target)
	}

	// acquire a shutdown lock. we don't want these next calls to be interrupted
	lock := shutdown.Lock()
	if lock == nil {
		return nil
	}
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err != nil {
		return err
	}
	volume, err := repository.FindVolume(volumeID)
	if err != nil {
		return err
	}
	snapshot, err := knoxite.NewSnapshot(opts.Description)
	if err != nil {
		return err
	}
	chunkIndex, err := knoxite.OpenChunkIndex(&repository)
	if err != nil {
		return err
	}
	// release the shutdown lock
	lock()

	err = store(&repository, &chunkIndex, snapshot, targets, opts)
	if err != nil {
		return err
	}

	// acquire another shutdown lock. we don't want these next calls to be interrupted
	lock = shutdown.Lock()
	if lock == nil {
		return nil
	}
	defer lock()

	err = snapshot.Save(&repository)
	if err != nil {
		return err
	}
	err = volume.AddSnapshot(snapshot.ID)
	if err != nil {
		return err
	}
	err = chunkIndex.Save(&repository)
	if err != nil {
		return err
	}
	return repository.Save()
}
