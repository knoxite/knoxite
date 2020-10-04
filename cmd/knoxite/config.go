/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */
package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/knoxite/knoxite/cmd/knoxite/config"
	"github.com/muesli/gotable"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "manage configuration",
		Long:  `The config command manages the knoxite configuration`,
	}
	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: "initialize a new configuration",
		Long:  "The init command initializes a new configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeConfigInit()
		},
	}
	configAliasCmd = &cobra.Command{
		Use:   "alias <alias>",
		Short: "Set an alias for the storage backend url to a repository",
		Long:  `The set command adds an alias for the storage backend url to a repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("alias needs an ALIAS to set")
			}
			return executeConfigAlias(args[0])
		},
	}
	configSetCmd = &cobra.Command{
		Use:   "set <option> <value>",
		Short: "set configuration values for an alias",
		Long:  "The set command lets you set configuration values for an alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("set needs to know which option to set")
			}
			if len(args) < 2 {
				return fmt.Errorf("set needs to know which value to set")
			}
			return executeConfigSet(args[0], args[1:])
		},
	}
	configInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "display information about the configuration file on stdout",
		Long:  `The info command displays information about the configuration file on stdout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeConfigInfo()
		},
	}
	configCatCmd = &cobra.Command{
		Use:   "cat",
		Short: "display the configuration file on stdout",
		Long:  `The cat command displays the configuration file on stdout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeConfigCat()
		},
	}
	configConvertCmd = &cobra.Command{
		Use:   "convert <source> <target>",
		Short: "convert between several configuration backends",
		Long:  "The convert command translates between several configuration backends",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("convert needs a source to work on")
			}
			if len(args) < 2 {
				return fmt.Errorf("convert needs a target to write to")
			}
			return executeConfigConvert(args[0], args[1])
		},
	}
)

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configAliasCmd)
	configCmd.AddCommand(configInfoCmd)
	configCmd.AddCommand(configCatCmd)
	configCmd.AddCommand(configConvertCmd)
	RootCmd.AddCommand(configCmd)
}

func executeConfigInit() error {
	log.Printf("Writing configuration file to: %s\n", cfg.URL().Path)
	err := cfg.Save()
	if err != nil {
		return err
	}
	logger.Infof("Wrote configuration file to: %s", cfg.URL().Path)
	return nil
}

func executeConfigAlias(alias string) error {
	// At first check if the configuration file already exists
	logger.Debug("Adding alias to config")
	cfg.Repositories[alias] = config.RepoConfig{
		Url: globalOpts.Repo,
		// Compression: utils.CompressionText(knoxite.CompressionNone),
		// Tolerance:   0,
		// Encryption:  utils.EncryptionText(knoxite.EncryptionAES),
	}

	logger.Info("Saving config")
	err := cfg.Save()
	if err != nil {
		return err
	}
	logger.Info("Saved config")
	return nil
}

func executeConfigSet(option string, values []string) error {
	// This probably won't scale for more complex configuration options but works
	// fine for now.
	parts := strings.Split(option, ".")
	if len(parts) != 2 {
		return fmt.Errorf("config set needs to work on an alias and a option like this: alias.option")
	}

	// The first part should be the repos alias
	logger.Info("Looking up repository config")
	repo, ok := cfg.Repositories[strings.ToLower(parts[0])]
	if !ok {
		return fmt.Errorf("No alias with name %s found", parts[0])
	}
	logger.Infof("Found repository configuration for alias %s", parts[0])

	logger.Info("Setting config options according to flags")
	opt := strings.ToLower(parts[1])
	switch opt {
	case "url":
		repo.Url = values[0]
	case "compression":
		repo.Compression = values[0]
	case "encryption":
		repo.Encryption = values[0]
	case "tolerance":
		tol, err := strconv.Atoi(values[0])
		if err != nil {
			return fmt.Errorf("Failed to convert %s to uint for the fault tolerance option: %v", opt, err)
		}
		repo.Tolerance = uint(tol)
	case "store_excludes":
		repo.StoreExcludes = values
	case "restore_excludes":
		repo.RestoreExcludes = values
	case "pedantic":
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		repo.Pedantic = b

	default:
		return fmt.Errorf("Unknown configuration option: %s", opt)
	}
	logger.Info("Set config options")

	cfg.Repositories[strings.ToLower(parts[0])] = repo

	logger.Info("Saving config")
	err := cfg.Save()
	if err != nil {
		return err
	}
	logger.Info("Saved config")
	return nil
}

func executeConfigInfo() error {
	logger.Debug("Initialising new gotable for output")
	tab := gotable.NewTable(
		[]string{"Alias", "Storage URL", "Compression", "Tolerance", "Encryption"},
		[]int64{-15, -35, -15, -15, 15},
		"No repository configurations found.")

	logger.Debug("Iterating over repositories to print details")
	for alias, repo := range cfg.Repositories {
		tab.AppendRow([]interface{}{
			alias,
			repo.Url,
			repo.Compression,
			fmt.Sprintf("%v", repo.Tolerance),
			repo.Encryption,
		})
	}
	return tab.Print()
}

func executeConfigCat() error {
	buf := new(bytes.Buffer)
	logger.Info("Marshalling toml config")
	if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
		return err
	}
	logger.Info("Marshalled toml config")

	fmt.Printf("%s\n", buf)
	return nil
}

func executeConfigConvert(source string, target string) error {
	// Load the source config
	logger.Info("Loading source config")
	logger.Info("Creating new config struct from source")
	scr, err := config.New(source)
	if err != nil {
		return err
	}
	logger.Info("Created new config")

	logger.Info("Loading source config")
	if err = scr.Load(); err != nil {
		return err
	}
	logger.Info("Loaded source config")

	logger.Info("Creating empty target config")
	// Create the target
	tar, err := config.New(target)
	if err != nil {
		return err
	}
	logger.Info("Created target config")

	logger.Info("Copying repo configs from source to target")
	// copy over the repo configs and save the target
	tar.Repositories = scr.Repositories

	logger.Info("Saving target config")
	err = tar.Save()
	if err != nil {
		return err
	}
	logger.Info("Saved target config")
	logger.Debug("You may now delete your old config")

	return nil
}
