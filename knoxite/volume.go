package main

import (
	"errors"
	"fmt"

	"knoxite"
)

// CmdVolume describes the command
type CmdVolume struct {
	Description string `short:"d" long:"desc" description:"a description or comment for this volume"`

	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("volume",
		"manage volumes",
		"The volume command manages volumes",
		&CmdVolume{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdVolume) Usage() string {
	return "[list|init]"
}

// Execute this command
func (cmd CmdVolume) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}

	switch args[0] {
	case "init":
		if len(args) < 2 {
			return fmt.Errorf(TWrongNumArgs, cmd.Usage())
		}
		return cmd.init(args[1])
	case "list":
		return cmd.list()
	}

	return nil
}

func (cmd CmdVolume) init(name string) error {
	repository, err := knoxite.OpenRepository(cmd.global.Repo, cmd.global.Password)
	if err == nil {
		vol, verr := knoxite.NewVolume(name, cmd.Description)
		if verr == nil {
			verr = repository.AddVolume(vol)
			if verr != nil {
				return fmt.Errorf("Creating volume %s failed: %v", name, verr)
			}

			fmt.Printf("Volume %s (Name: %s, Description: %s) created\n", vol.ID, name, cmd.Description)
			return repository.Save()
		}
	}
	return err
}

func (cmd CmdVolume) list() error {
	repository, err := knoxite.OpenRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return err
	}

	tab := NewTable([]string{"ID", "Name", "Description"}, []int64{-8, -32, -48}, "No volumes found. This repository is empty.")
	for _, volume := range repository.Volumes {
		tab.Rows = append(tab.Rows, []interface{}{volume.ID, volume.Name, volume.Description})
	}

	tab.Print()

	return nil
}
