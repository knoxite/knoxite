/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/knoxite/knoxite/config"
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
			return executeConfigSet(args[0], args[1])
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
	return cfg.Save()
}

func executeConfigAlias(alias string) error {
	// At first check if the configuration file already exists
	cfg.Repositories[alias] = config.RepoConfig{
		Url: globalOpts.Repo,
		// Compression: utils.CompressionText(knoxite.CompressionNone),
		// Tolerance:   0,
		// Encryption:  utils.EncryptionText(knoxite.EncryptionAES),
	}

	return cfg.Save()
}

func executeConfigSet(option string, value string) error {
	// This probably wont scale for more complex configuration options but works
	// fine for now.
	parts := strings.Split(option, ".")
	if len(parts) != 2 {
		return fmt.Errorf("config set needs to work on an alias and a option like this: alias.option")
	}

	// The first part should be the repos alias
	repo, ok := cfg.Repositories[strings.ToLower(parts[0])]
	if !ok {
		return fmt.Errorf("No alias with name %s found", parts[0])
	}

	opt := strings.ToLower(parts[1])
	switch opt {
	case "url":
		repo.Url = value
	case "compression":
		repo.Compression = value
	case "encryption":
		repo.Encryption = value
	case "tolerance":
		tol, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("Failed to convert %s to uint for the fault tolerance option: %v", opt, err)
		}
		repo.Tolerance = uint(tol)
	default:
		return fmt.Errorf("Unknown configuration option: %s", opt)
	}
	cfg.Repositories[strings.ToLower(parts[0])] = repo

	return cfg.Save()
}

func executeConfigInfo() error {
	tab := gotable.NewTable(
		[]string{"Alias", "Storage URL", "Compression", "Tolerance", "Encryption"},
		[]int64{-15, -35, -15, -15, 15},
		"No repository configurations found.")

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
	json, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", json)
	return nil
}

func executeConfigConvert(source string, target string) error {
	// Load the source config
	scr, err := config.New(source)
	if err != nil {
		return err
	}
	if err = scr.Load(); err != nil {
		return err
	}

	// Create the target
	tar, err := config.New(target)
	if err != nil {
		return err
	}

	// copy over the repo configs and save the target
	tar.Repositories = scr.Repositories
	return tar.Save()
}
