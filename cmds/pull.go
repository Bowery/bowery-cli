// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/prompt"
	"github.com/Bowery/cli/requests"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/tar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/schemas"
)

func init() {
	Cmds["pull"] = &Cmd{pullRun, "pull [id or name]", "Pull down an application and its code.", ""}
}

func pullRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := flag.Lookup("force").Value.String()

	// If there is more than 1 argument, exit.
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["pull"].Usage, "\n\n"+Cmds["pull"].Short)
		return 2
	}

	// Get developer
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If there's no developer get a token.
	if dev.Token == "" {
		log.Println("yellow", "Oops! You must be logged in to perform this action.")
		return 0
	}

	inAppDir := false
	isAppOwner := false

	var app *schemas.Application
	var state *db.State

	// If there is no argument passed ensure that we are in an application
	// directory.
	if len(args) == 0 {
		state, err = db.GetState()
		if err == errors.ErrNotConnected {
			fmt.Fprintln(os.Stderr,
				"Usage: bowery "+Cmds["pull"].Usage, "\n\n"+Cmds["pull"].Description)
			return 2
		}
		if err != nil {
			rollbar.Report(err)
			return 1
		}
		if state.App.ID != "" {
			log.Debug("Inside application directory with ID " + state.App.ID)
			app = state.App
			inAppDir = true
		}
	}

	// Get app
	if !inAppDir {
		log.Debug("Getting app with id:", args[0])
		app, err = requests.GetAppById(args[0])
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	// If the app is owned by the active developer
	// set isAppOwner true.
	if dev.Developer.ID == app.DeveloperID {
		log.Debug("Current developer owns application.")
		isAppOwner = true
	}

	log.Debug("Found app successfully")

	dir, err := os.Getwd()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	// If the destination is not-empty confirm with the user that the
	// action will overwrite the contents of the current directory.
	if len(files) > 0 && force != "true" {
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
	var services db.Services
	services.Data = make(map[string]*db.Service)
	if inAppDir {
		services.Data = state.Config
	}
	for _, service := range app.Services {
		if !inAppDir {
			ports := make([]interface{}, 0, len(service.CustomPorts))
			for port := range service.CustomPorts {
				portNum, _ := strconv.Atoi(port)
				ports = append(ports, portNum)
			}

			services.Data[service.Name] = &db.Service{
				Image: service.Image,
				Path:  filepath.Join(dir, service.Name),
				Ports: ports,
				Start: service.Start,
				Build: service.Build,
				Test:  service.Test,
			}
		}

		// Get new files and write to appropriate directory. If working
		// within a pre-existing application write to appropriate
		// paths based on active bowery.json file.
		var path string
		if !inAppDir {
			path = filepath.Join(dir, service.Name)
			err = os.MkdirAll(path, 0777)
			if err != nil {
				rollbar.Report(err)
				return 1
			}
		} else {
			if services.Data[service.Name].Path == "" {
				path = ""
			} else {
				path = filepath.Join(dir, services.Data[service.Name].Path)
			}
		}

		if path != "" {
			log.Debug("Fetching " + service.Name + " files")
			res, err := http.Get("http://" + service.SatelliteAddr)
			if err != nil {
				rollbar.Report(errors.NewStackError(err))
				return 1
			}

			// If 200 untar files to directory.
			if res.StatusCode == 200 {
				log.Debug("Writing", service.Name, "files to", path)
				err = tar.Untar(res.Body, path)
				if err != nil {
					rollbar.Report(err)
					return 1
				}
			} else {
				log.Debug("Unable to get files for ", service.Name)
			}
		}
	}

	// Generate state if is app owner.
	if isAppOwner {
		state := &db.State{
			Token:  dev.Token,
			App:    app,
			Config: services.Data,
			Path:   filepath.Join(dir, ".bowery", "state"),
		}
		err = state.Save()
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	// Generate bowery.json.
	if !inAppDir {
		configPath, err := os.Getwd()
		services.Path = filepath.Join(configPath, "bowery.json")
		if err != nil {
			return 1
		}
		err = services.Save()
		if err != nil {
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
