package main

import (
	"errors"
	"fmt"

	"github.com/knoxite/knoxite"
)

// CmdRepository describes the command
type CmdRepository struct {
	global *GlobalOptions
}

func init() {
	_, err := parser.AddCommand("repo",
		"manage repository",
		"The repo command manages repositories",
		&CmdRepository{global: &globalOpts})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdRepository) Usage() string {
	return "[init]"
}

// Execute this command
func (cmd CmdRepository) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return errors.New(TSpecifyRepoLocation)
	}

	switch args[0] {
	case "init":
		return cmd.init()
	}

	return nil
}

func (cmd CmdRepository) init() error {
	/*	username := ""
		user, err := user.Current()
		if err == nil {
			username = user.Username
		}
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}*/

	_, err := knoxite.NewRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return fmt.Errorf("Creating repository at %s failed: %v", cmd.global.Repo, err)
	}

	fmt.Printf("Created new repository at %s\n", cmd.global.Repo)
	return nil
}
