package action

import (
	"strings"

	"github.com/rsteube/carapace"
)

func hasPathPrefix(s string) bool {
	return strings.HasPrefix(s, ".") ||
		strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "~")
}

func ActionRepo() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		if hasPathPrefix(c.CallbackValue) {
			return carapace.ActionDirectories()
		}
		return carapace.ActionMultiParts("://", func(c carapace.Context) carapace.Action {
			switch len(c.Parts) {
			case 0:
				return carapace.ActionValuesDescribed(
					"azurefile", "Azure File Storage",
					"backblaze", "backblaze",
					"dropbox", "Dropbox",
					"ftp", "FTP",
					"googlecloudstorage", "Google Cloud Storage",
					"mega", "Mega",
					"s3", "Amazon S3",
					"s3s", "Amazon S3 SSL",
					"sftp", "SSH/SFTP",
					"webdavs", "WebDAV",
				).Invoke(c).Suffix("://").ToA()
			default:
				return carapace.ActionValues()
			}
		})
	})
}
