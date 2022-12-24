package action

import (
	"fmt"

	"github.com/knoxite/knoxite"
	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"
	"github.com/spf13/cobra"
)

func ActionVolumes(cmd *cobra.Command) carapace.Action {
	return actionRepository(cmd, func(repository knoxite.Repository) carapace.Action {
		vals := make([]string, 0)
		for _, volume := range repository.Volumes {
			if volume.Description == "" {
				vals = append(vals, volume.ID, volume.Name)
			} else {
				vals = append(vals, volume.ID, fmt.Sprintf("%v - %v", volume.Name, volume.Description))
			}
		}
		return carapace.ActionValuesDescribed(vals...).Style(style.Blue)
	}).Tag("volumes")
}
