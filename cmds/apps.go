// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
)

func init() {
	Cmds["apps"] = &Cmd{
		Run:   appsRun,
		Usage: "apps",
		Short: "List all of your Bowery Apps.",
	}
}

func appsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := getDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	// Fetch apps.
	log.Print("magenta", "Requesting apps... ")
	apps, err := api.GetApps(dev.Token)
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
