// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"strings"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["logout"] = &Cmd{logoutRun, "logout", "Log out of your Bowery account.", ""}
}

func logoutRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If we have a token, we're logged in.
	if dev.Token != "" {
		log.Println("", "Logging you out",
			strings.Split(dev.Developer.Name, " ")[0]+".")
	} else {
		log.Println("yellow", "No user logged in.")
	}

	dev.Token = ""
	dev.Developer = nil
	err = dev.Save()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery logout", map[string]*db.Developer{"user": dev})

	return 0
}
