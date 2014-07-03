// Copyright 2013-2014 Bowery, Inc
package cmds

import (
	"flag"
	"fmt"
	"os"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["destroy"] = &Cmd{destroyRun, "destroy <id or name>", "Destroy an application and its services.", ""}
}

func destroyRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := flag.Lookup("force").Value.String()
	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["destroy"].Usage, "\n\n"+Cmds["destroy"].Short)
		return 2 // --help uses 2.
	}

	// Get developer.
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	if dev.Token == "" {
		log.Println("yellow", "Oops! You must be logged in to perform this action.")
		return 1
	}

	// Fetch app.
	log.Debug("Getting app with id:", args[0])
	app, err := requests.GetAppById(args[0])
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Found app successfully:", app.ID)

	// Make sure developer owns app.
	if dev.Developer.ID == app.DeveloperID {
		log.Debug("Current developer owns application.")
	} else {
		log.Println("yellow", "You must be the owner of this application to perform this action.")
		return 1
	}

	// Ask for confirmation to delete.
	if force != "true" {
		ok, err := prompt.Ask("Are you sure you want to delete this app")
		if err != nil {
			rollbar.Report(err)
			return 1
		}
		if !ok {
			log.Println("yellow", "Application NOT removed.")
			return 0
		}
	}

	// Remove services.
	for _, service := range app.Services {
		log.Println("yellow", "Removing service", service.Name)
		if err := requests.RemoveService(service.DockerID, dev.Token); err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	// Send delete requests.
	log.Println("yellow", "Attempting to remove application...")
	if err := requests.DestroyAppByID(app.ID, dev.Token); err != nil {
		rollbar.Report(err)
		return 1
	}

	state, err := db.GetState()
	if err != nil && err != errors.ErrNotConnected {
		rollbar.Report(err)
		return 1
	}

	if err == nil {
		// Remove app state if the id matches the current application.
		if state.App.ID == app.ID {
			if err := os.RemoveAll(".bowery"); err != nil {
				rollbar.Report(err)
				return 1
			}
		}
	}

	keen.AddEvent("bowery destroy", app)
	log.Println("yellow", "Application removed sucessfully.")
	return 0
}
