/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020,      Nicolas Martin <penguwin@penguwin.eu>
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

	shutdown "github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite/cfg"
	_ "github.com/knoxite/knoxite/storage/azure"
	_ "github.com/knoxite/knoxite/storage/backblaze"
	_ "github.com/knoxite/knoxite/storage/dropbox"
	_ "github.com/knoxite/knoxite/storage/ftp"
	_ "github.com/knoxite/knoxite/storage/googlecloud"
	_ "github.com/knoxite/knoxite/storage/http"
	_ "github.com/knoxite/knoxite/storage/mega"
	_ "github.com/knoxite/knoxite/storage/s3"
	_ "github.com/knoxite/knoxite/storage/sftp"
	_ "github.com/knoxite/knoxite/storage/webdav"
)

// GlobalOptions holds all those options that can be set for every command
type GlobalOptions struct {
	Repo      string
	Password  string
	ConfigURL string
}

var (
	Version   = ""
	CommitSHA = ""

	globalOpts = GlobalOptions{}
	config     = &cfg.Config{}

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
	RootCmd.PersistentFlags().StringVarP(&globalOpts.ConfigURL, "configURL", "C", cfg.DefaultPath(), "Path to the configuration file")

	globalOpts.Repo = os.Getenv("KNOXITE_REPOSITORY")
	globalOpts.Password = os.Getenv("KNOXITE_PASSWORD")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	if CommitSHA != "" {
		vt := RootCmd.VersionTemplate()
		RootCmd.SetVersionTemplate(vt[:len(vt)-1] + " (" + CommitSHA + ")\n")
	}
	if Version == "" {
		Version = "unknown (built from source)"
	}

	RootCmd.Version = Version
}

// initConfig initializes the configuration for knoxite.
// It'll use the the default config url unless specified otherwise via the
// ConfigURL flag.
func initConfig() {
	var err error
	config, err = cfg.New(globalOpts.ConfigURL)
	if err != nil {
		log.Fatalf("error reading the config file: %v\n", err)
		return
	}
	if err = config.Load(); err != nil {
		log.Fatalf("error loading the config file: %v\n", err)
		return
	}

	// There can occur a panic due to an entry assigment in nil map when theres
	// no map initialized to store the RepoConfigs. This will prevent this from
	// happening:
	if config.Repositories == nil {
		config.Repositories = make(map[string]cfg.RepoConfig)
	}
}
