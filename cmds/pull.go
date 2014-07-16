// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/delancey"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
	"github.com/Bowery/gopackages/schemas"
	"github.com/Bowery/gopackages/tar"
)

func init() {
	Cmds["pull"] = &Cmd{
		Run:   pullRun,
		Usage: "pull [id or name]",
		Short: "Pull down an application and its code.",
	}
}

func pullRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := Cmds["pull"].Force

	dev, err := getDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	var (
		app   *schemas.Application
		state *db.State
	)
	inAppDir := false
	isAppOwner := false

	// If there is no argument passed ensure that we are in an application
	// directory.
	if len(args) <= 0 {
		state, err = db.GetState()
		if err != nil && err != errors.ErrNotConnected {
			rollbar.Report(err)
			return 1
		}

		if err == errors.ErrNotConnected || state.App.ID == "" {
			fmt.Fprintln(os.Stderr,
				"Usage: bowery "+Cmds["pull"].Usage, "\n\n"+Cmds["pull"].Short)
			return 2
		}

		log.Debug("Inside application directory with ID " + state.App.ID)
		app = state.App
		inAppDir = true
	}

	// Get app
	if !inAppDir {
		log.Debug("Getting app with id:", args[0])
		app, err = api.GetAppById(args[0])
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	if dev.Developer.ID.Hex() == app.DeveloperID {
		log.Debug("Current developer owns application.")
		isAppOwner = true
	}

	log.Debug("Found app successfully")
	file, err := os.Open(".")
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	files, err := file.Readdir(0)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	file.Close()

	// If the destination is not-empty confirm with the user that the
	// action will overwrite the contents of the current directory.
	if len(files) > 0 && !force {
		log.Println("magenta", "Looks like you're trying to pull down an app into a non-empty directory")
		log.Println("magenta", "Proceeding will overwrite the contents of this directory.")
		ok, err := prompt.Ask("Are you sure you want to proceed")
		if err != nil {
			rollbar.Report(err)
			return 1
		}
		if !ok {
			return 0
		}
	}

	// Generate bowery.json and load service files.
	services := new(db.Services)
	if inAppDir {
		services.Data = state.Config
	} else {
		services.Data = make(map[string]*db.Service)
	}

	for _, service := range app.Services {
		// Add service to bowery.json list.
		if !inAppDir {
			ports := make([]interface{}, 0, len(service.CustomPorts))
			for port := range service.CustomPorts {
				portNum, _ := strconv.Atoi(port)
				ports = append(ports, portNum)
			}

			services.Data[service.Name] = &db.Service{
				Image: service.Image,
				Path:  service.Name,
				Ports: ports,
				Start: service.Start,
				Build: service.Build,
				Test:  service.Test,
			}
		}

		// Create the path for the service, and download its files.
		path := services.Data[service.Name].Path
		if path != "" {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				rollbar.Report(err)
				return 1
			}

			log.Debug("Fetching", service.Name, "files")
			contents, err := delancey.Download(service.SatelliteAddr)
			if err != nil {
				rollbar.Report(err)
				return 1
			}

			err = tar.Untar(contents, path)
			if err != nil {
				rollbar.Report(err)
				return 1
			}
		}
	}

	// Generate state if is app owner.
	if isAppOwner {
		state := &db.State{
			Token:  dev.Token,
			App:    app,
			Config: services.Data,
			Path:   filepath.Join(".bowery", "state"),
		}
		err = state.Save()
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	// Generate bowery.json.
	if !inAppDir {
		services.Path = "bowery.json"
		err = services.Save()
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		// Prompt the developer to now run bowery connect.
		log.Println("magenta", "To get running just execute:\n")
		log.Println("magenta", "  $ bowery connect")
	}

	keen.AddEvent("bowery pull", map[string]interface{}{
		"user": dev,
		"app":  app,
	})
	return 0
}
