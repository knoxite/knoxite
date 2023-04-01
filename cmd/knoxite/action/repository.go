package action

import (
	"os"

	"github.com/knoxite/knoxite"
	"github.com/knoxite/knoxite/cmd/knoxite/config"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

func actionRepository(cmd *cobra.Command, f func(repository knoxite.Repository) carapace.Action) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		repo := os.Getenv("KNOXITE_REPOSITORY")
		password := os.Getenv("KNOXITE_PASSWORD")
		if f := cmd.Flag("repo"); f.Changed {
			repo = f.Value.String()
		}
		if f := cmd.Flag("password"); f.Changed {
			password = f.Value.String()
		}

		// We dont allow both flags to be set as this can lead to unclear instructions.
		if cmd.Flags().Changed("repo") && cmd.Flags().Changed("alias") {
			return carapace.ActionMessage("Specify either repository directory '-r' or an alias '-R'")
		}

		if f := cmd.Flag("alias"); f.Changed {
			cfg, err := config.New(cmd.Flag("configURL").Value.String())
			if err != nil {
				return carapace.ActionMessage("Error reading the config file: %v", err)
			}

			if err = cfg.Load(); err != nil {
				return carapace.ActionMessage("Error parsing the toml config file at '%s': %v", cfg.URL().Path, err)
			}

			// There can occur a panic due to an entry assigment in nil map when theres
			// no map initialized to store the RepoConfigs. This will prevent this from
			// happening:
			if cfg.Repositories == nil {
				cfg.Repositories = make(map[string]config.RepoConfig)
			}

			rep, ok := cfg.Repositories[f.Value.String()]
			if !ok {
				return carapace.ActionMessage("Error loading the specified alias")
			}
			repo = rep.Url
		}

		repository, err := knoxite.OpenRepository(repo, password)
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}
		return f(repository)
	})
}
