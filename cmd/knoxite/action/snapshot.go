package action

import (
	"fmt"

	"github.com/knoxite/knoxite"
	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"
	"github.com/spf13/cobra"
)

func ActionSnapshots(cmd *cobra.Command, volumeID string) carapace.Action {
	return actionRepository(cmd, func(repository knoxite.Repository) carapace.Action {
		vals := make([]string, 0)
		for _, volume := range repository.Volumes {
			if volumeID == "" || volume.ID == volumeID {
				description := volume.Name
				if volume.Description != "" {
					description = fmt.Sprintf("%v - %v", volume.Name, volume.Description)
				}
				for _, snapshot := range volume.Snapshots {
					vals = append(vals, snapshot, description)
				}
			}
		}
		return carapace.ActionValuesDescribed(vals...).Style(style.Yellow)
	}).Tag("snapshots")
}

func ActionSnapshotPaths(cmd *cobra.Command, snapshotID string) carapace.Action {
	return actionRepository(cmd, func(repository knoxite.Repository) carapace.Action {
		_, snapshot, err := repository.FindSnapshot(snapshotID)
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}

		vals := make([]string, 0)
		for _, archive := range snapshot.Archives {
			if archive.Type == knoxite.File {
				vals = append(vals, archive.Path)
			}
		}
		return carapace.ActionValues(vals...).StyleF(style.ForPathExt)
	}).Tag("snapshot paths")
}
