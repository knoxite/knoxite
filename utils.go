package knoxite

import (
	"os"
	"os/user"
)

func CurrentUser() string {
	usr, err := user.Current()
	if err != nil {
		return "unknown"
	}

	return usr.Username
}

func CurrentHost() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return h
}
