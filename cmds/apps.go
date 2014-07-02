// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["apps"] = &Cmd{appsRun, "apps", "List all of your Bowery Apps.", ""}
}

func appsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If there's no developer get a token.
	if dev.Token == "" {
		log.Println("yellow", "Oops! You must be logged in.")
		return 1
	}

	// Fetch apps.
	log.Print("magenta", "Requesting apps... ")
	apps, err := requests.GetApps(dev.Token)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Founds apps", apps)

	if len(apps) <= 0 {
		log.Println("", "No apps were found.")
	} else {
		log.Println("magenta", "Found", len(apps), "app(s).\n")
		for _, app := range apps {
			if app.Name != "" {
				log.Println("", app.Name)
			} else {
				log.Println("", app.ID)
			}
		}
	}

	log.Println("magenta", "\nIf you'd like to start working on one of these apps run:\n")
	log.Println("magenta", "  $ bowery pull <app name or id>\n")
	keen.AddEvent("bowery apps", map[string]*db.Developer{"user": dev})
	return 0
}
