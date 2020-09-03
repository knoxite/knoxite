/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/renderers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type VerifyOptions struct {
	Percentage int
	Output     knoxite.DefaultOutput
}

var (
	verifyOpts = VerifyOptions{}

	verifyCmd = &cobra.Command{
		Use:   "verify [<volume> [<snapshot>]]",
		Short: "verify a repo, volume or snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return executeVerifyRepo(verifyOpts)
			} else if len(args) == 1 {
				return executeVerifyVolume(args[0], verifyOpts)
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

func executeVerifyRepo(opts VerifyOptions) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()
	errors := make([]error, 0)

	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		progress, err := knoxite.VerifyRepo(repository, opts.Percentage)
		if err != nil {
			errors = append(errors, err)
			return err
		}

		verifyRenderer := renderers.VerifyRenderer{
			Errors:         &errors,
			DisposeMessage: fmt.Sprintf("Verify done: %d errors", len(errors)),
		}

		output := knoxite.DefaultOutput{
			Renderers: knoxite.Renderers{&verifyRenderer},
		}

		err = output.Init()
		if err != nil {
			return nil
		}

		err = output.Render(progress, cancel)
		if err != nil {
			return err
		}

		return output.Dispose()
	}
	return err
}

func executeVerifyVolume(volumeId string, opts VerifyOptions) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()
	errors := make([]error, 0)

	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		progress, err := knoxite.VerifyVolume(repository, volumeId, opts.Percentage)
		if err != nil {
			errors = append(errors, err)
			return err
		}

		verifyRenderer := renderers.VerifyRenderer{
			Errors:         &errors,
			DisposeMessage: fmt.Sprintf("Verify done: %d errors", len(errors)),
		}

		output := knoxite.DefaultOutput{
			Renderers: knoxite.Renderers{&verifyRenderer},
		}

		err = output.Init()
		if err != nil {
			return nil
		}

		err = output.Render(progress, cancel)
		if err != nil {
			return err
		}

		return output.Dispose()
	}
	return err
}

func executeVerifySnapshot(volumeId string, snapshotId string, opts VerifyOptions) error {
	// we want to be notified during the first phase of a shutdown
	cancel := shutdown.First()
	errors := make([]error, 0)

	repository, err := openRepository(globalOpts.Repo, globalOpts.Password)
	if err == nil {
		progress, err := knoxite.VerifySnapshot(repository, snapshotId, opts.Percentage)
		if err != nil {
			errors = append(errors, err)
			return err
		}

		verifyRenderer := renderers.VerifyRenderer{
			Errors:         &errors,
			DisposeMessage: fmt.Sprintf("Verify done: %d errors", len(errors)),
		}

		output := knoxite.DefaultOutput{
			Renderers: knoxite.Renderers{&verifyRenderer},
		}

		err = output.Init()
		if err != nil {
			return nil
		}

		err = output.Render(progress, cancel)
		if err != nil {
			return err
		}

		return output.Dispose()
	}
	return err
}
