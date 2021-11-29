package action

import (
	"fmt"

	"github.com/knoxite/knoxite/cmd/knoxite/config"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

func ActionAliases(cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		cfg, err := config.New(cmd.Flag("configURL").Value.String())
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}

		if err = cfg.Load(); err != nil {
			return carapace.ActionMessage(fmt.Sprintf("Error parsing the toml config file at '%s': %v", cfg.URL().Path, err))
		}

		vals := make([]string, 0)
		for alias, repo := range cfg.Repositories {
			vals = append(vals, alias, repo.Url)
		}
		return carapace.ActionValuesDescribed(vals...)
	})
}

func ActionConfigKeys(cmd *cobra.Command) carapace.Action {
	return carapace.ActionMultiParts(".", func(c carapace.Context) carapace.Action {
		switch len(c.Parts) {
		case 0:
			return ActionAliases(cmd).Invoke(c).Suffix(".").ToA()
		case 1:
			return carapace.ActionValuesDescribed(
				"url", "Repository directory to backup to/restore from",
				"compression", "Compression algo to use: none (default), flate, gzip, lzma, zlib, zstd",
				"tolerance", "Failure tolerance against n backend failures",
				"encryption", "Encryption algo to use: aes (default), none",
				"pedantic", "Stop backup operation after the first error occurred",
				"store_excludes", "Specify excludes for the store operation",
				"restore_excludes", "Specify excludes for the restore operation",
			)
		default:
			return carapace.ActionValues()
		}
	})
}

func ActionConfigValues(key string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		switch key {
		case "url":
			return ActionRepo()
		case "compression":
			return carapace.ActionValues("none", "flate", "gzip", "lzma", "zlib", "zstd")
		case "encryption":
			return carapace.ActionValues("aes", "none")
		case "pedantic":
			return carapace.ActionValues("true", "false")
		default:
			return carapace.ActionValues()
		}
	})
}
