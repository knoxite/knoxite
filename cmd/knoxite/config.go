/*
 * knoxite
 *     Copyright (c) 2020, Nicolas Martin <penguwin@penguwin.eu>
 *
 *   For license see LICENSE
 */
package main

import (
	"log"

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
)

func init() {
	configCmd.AddCommand(configInitCmd)
	RootCmd.AddCommand(configCmd)
}

func executeConfigInit() error {
	log.Printf("Writing configuration file to: %s\n", config.URL().Path)
	return config.Save()
}
