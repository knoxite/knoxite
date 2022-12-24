/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2020-2021, Nicolas Martin <penguwin@penguwin.eu>
 *     Copyright (c) 2020,      Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"os"
	"syscall"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"
	"github.com/spf13/cobra"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/action"
	"github.com/knoxite/knoxite/cmd/knoxite/config"
	"github.com/knoxite/knoxite/cmd/knoxite/utils"
	_ "github.com/knoxite/knoxite/storage/amazons3"
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

// GlobalOptions holds all those options that can be set for every command.
type GlobalOptions struct {
	Repo      string
	Alias     string
	Password  string
	ConfigURL string
	Verbose   int
	LogLevel  string
}

var (
	Version   = ""
	CommitSHA = ""

	globalOpts = GlobalOptions{}
	cfg        = &config.Config{}

	// RootCmd is the core command used for cli-arg parsing.
	RootCmd = &cobra.Command{
		Use:   "knoxite",
		Short: "Knoxite is a data storage & backup tool",
		Long: "Knoxite is a secure and flexible data storage and backup tool\n" +
			"Complete documentation is available at https://github.com/knoxite/knoxite",
		SilenceErrors:     true,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	log knoxite.Logger
)

func main() {
	shutdown.OnSignal(0, os.Interrupt, syscall.SIGTERM)
	// use quiet knoxite.NopLogger as shutdown logger
	shutdown.Logger = shutdown.LogPrinter(knoxite.NopLogger{})
	// shutdown.SetTimeout(0)

	RootCmd.PersistentFlags().StringVarP(&globalOpts.Repo, "repo", "r", "", "Repository directory to backup to/restore from (default: current working dir)")
	RootCmd.PersistentFlags().StringVarP(&globalOpts.Alias, "alias", "R", "", "Repository alias to backup to/restore from")
	RootCmd.PersistentFlags().StringVar(&globalOpts.Password, "password", "", "Password to use for data encryption")
	RootCmd.PersistentFlags().StringVarP(&globalOpts.ConfigURL, "configURL", "C", config.DefaultPath(), "Path to the configuration file")
	RootCmd.PersistentFlags().StringVar(&globalOpts.LogLevel, "loglevel", "Print", "Verbose output. Possible levels are Debug, Info, Warning and Fatal")
	RootCmd.PersistentFlags().CountVarP(&globalOpts.Verbose, "verbose", "v", "Verbose output on log level Info (-v) or Debug (-vv). Use --loglevel to choose between Debug, Info, Warning and Fatal")

	globalOpts.Repo = os.Getenv("KNOXITE_REPOSITORY")
	globalOpts.Password = os.Getenv("KNOXITE_PASSWORD")

	// add the `completion` command via carapace
	carapace.Gen(RootCmd).FlagCompletion(carapace.ActionMap{
		"alias":     action.ActionAliases(RootCmd),
		"repo":      action.ActionRepo(),
		"configURL": carapace.ActionFiles(),
		"loglevel":  carapace.ActionValues("Debug", "Info", "Warning", "Fatal").StyleF(style.ForLogLevel),
	})

	carapace.Override(carapace.Opts{
		BridgeCompletion: true,
	})

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initLogger)
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

func initLogger() {
	switch {
	case globalOpts.Verbose == 1:
		globalOpts.LogLevel = "Info"
	case globalOpts.Verbose >= 2:
		globalOpts.LogLevel = "Debug"
	}

	logLevel, err := utils.LogLevelFromString(globalOpts.LogLevel)

	log = *NewLogger(logLevel).
		WithWriter(os.Stdout)

	if err != nil {
		log.Warnf("Error setting log level \"%s\": %s. Using default log level Info instead.", globalOpts.LogLevel, err)
	}

	// set logger for knoxite lib
	knoxite.SetLogger(log)
}

// initConfig initializes the configuration for knoxite.
// It'll use the the default config url unless specified otherwise via the
// ConfigURL flag.
func initConfig() {
	// We dont allow both flags to be set as this can lead to unclear instructions.
	if RootCmd.PersistentFlags().Changed("repo") && RootCmd.PersistentFlags().Changed("alias") {
		log.Fatalf("Specify either repository directory '-r' or an alias '-R'")
		return
	}

	var err error
	cfg, err = config.New(globalOpts.ConfigURL)
	if err != nil {
		log.Fatalf("Error reading the config file: %v", err)
		return
	}

	if err = cfg.Load(); err != nil {
		log.Fatalf("Error parsing the toml config file at '%s': %v", cfg.URL().Path, err)
		return
	}

	// There can occur a panic due to an entry assigment in nil map when theres
	// no map initialized to store the RepoConfigs. This will prevent this from
	// happening:
	if cfg.Repositories == nil {
		cfg.Repositories = make(map[string]config.RepoConfig)
	}

	if globalOpts.Alias != "" {
		rep, ok := cfg.Repositories[globalOpts.Alias]
		if !ok {
			log.Fatalf("Error loading the specified alias")
			return
		}

		globalOpts.Repo = rep.Url
	}
}
