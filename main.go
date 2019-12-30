/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"

	_ "github.com/knoxite/knoxite/storage/amazons3"
	_ "github.com/knoxite/knoxite/storage/backblaze"
	_ "github.com/knoxite/knoxite/storage/dropbox"
	_ "github.com/knoxite/knoxite/storage/ftp"
	_ "github.com/knoxite/knoxite/storage/http"
	_ "github.com/knoxite/knoxite/storage/sftp"
)

// GlobalOptions holds all those options that can be set for every command
type GlobalOptions struct {
	Repo     string
	Password string
}

var (
	globalOpts = GlobalOptions{}

	// RootCmd is the core command used for cli-arg parsing
	RootCmd = &cobra.Command{
		Use:   "knoxite",
		Short: "Knoxite is a data storage & backup tool",
		Long: "Knoxite is a secure and flexible data storage and backup tool\n" +
			"Complete documentation is available at https://github.com/knoxite/knoxite",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func main() {
	shutdown.OnSignal(0, os.Interrupt, syscall.SIGTERM)
	// quiet shutdown logger
	shutdown.Logger = shutdown.LogPrinter(log.New(ioutil.Discard, "", log.LstdFlags))
	// shutdown.SetTimeout(0)

	RootCmd.PersistentFlags().StringVarP(&globalOpts.Repo, "repo", "r", "", "Repository directory to backup to/restore from (default: current working dir)")
	RootCmd.PersistentFlags().StringVarP(&globalOpts.Password, "password", "p", "", "Password to use for data encryption")

	globalOpts.Repo = os.Getenv("KNOXITE_REPOSITORY")
	globalOpts.Password = os.Getenv("KNOXITE_PASSWORD")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
