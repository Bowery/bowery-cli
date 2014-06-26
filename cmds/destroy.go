// Copyright 2013-2014 Bowery, Inc
package cmds

import (
	"flag"
	"fmt"
	"os"

	"github.com/Bowery/SkyLab/cli/db"
	"github.com/Bowery/SkyLab/cli/errors"
	"github.com/Bowery/SkyLab/cli/keen"
	"github.com/Bowery/SkyLab/cli/log"
	"github.com/Bowery/SkyLab/cli/prompt"
	"github.com/Bowery/SkyLab/cli/requests"
	"github.com/Bowery/SkyLab/cli/rollbar"
	"github.com/Bowery/SkyLab/cli/schemas"

	"labix.org/v2/mgo/bson"
)

func init() {
	Cmds["destroy"] = &Cmd{destroyRun, "destroy <id>", "Destroy an application and it's services.", ""}
}

func destroyRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := flag.Lookup("force").Value.String()

	if len(args) != 1 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["destroy"].Usage, "\n\n"+Cmds["destroy"].Description)
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
		return 0
	}

	// Fetch app.
	var app *schemas.Application
	appID := args[0]
	if isObjId := bson.IsObjectIdHex(appID); isObjId == true {
		log.Debug("Getting app with id:", appID)
		app, err = requests.GetAppById(appID)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	} else {
		log.Println("yellow", "A valid app id is required.")
		return 0
	}

	log.Debug("Found app successfully:", app.ID)

	// Make sure developer owns app.
	if dev.Developer.ID == app.DeveloperID {
		log.Debug("Current developer owns application.")
	} else {
		log.Println("yellow", "You must be the owner of this application to perform this action.")
		return 0
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
	if err := requests.DestroyAppByID(appID, dev.Token); err != nil {
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

	log.Println("yellow", "Application removed sucessfully.")

	return 0
}
