// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"strings"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/schemas"
)

func init() {
	Cmds["info"] = &Cmd{infoRun, "info", "Display developer info and application info.", ""}
}

func infoRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	state, err := db.GetState()
	if err != nil && err != errors.ErrNotConnected {
		rollbar.Report(err)
		return 1
	}

	if ENV == "development" {
		log.Println("", "Developer(development mode):")
	} else {
		log.Println("", "Developer:")
	}
	if dev.Token == "" {
		log.Println("yellow", "  Not logged in.")
	} else {
		log.Println("", "  Email:", dev.Developer.Email)
		log.Println("", "  Name:", dev.Developer.Name)
	}

	log.Println("", "\nConfig:")
	if dev.Config == nil || len(dev.Config) <= 0 {
		log.Println("yellow", "  No custom configuration.")
	} else {
		for k, v := range dev.Config {
			if v != "" {
				log.Println("", "  "+strings.Title(k)+":", v)
			}
		}
	}

	log.Println("", "\nApplication:")
	if err == errors.ErrNotConnected {
		log.Println("yellow", "  No connected application in this directory.")
	} else {
		log.Println("", "  ID:", state.App.ID)

		for name, service := range state.Config {
			// Get the associated api.Service from the app service list.
			var apiService *schemas.Service
			for _, serv := range state.App.Services {
				if serv.Name == name {
					apiService = serv
					break
				}
			}

			url := apiService.PublicAddr
			if ENV != "development" {
				url = "http://" + name + "." + state.App.ID + ".boweryapps.com"
			}

			log.Println("", "  "+name+":")
			log.Println("", "    URL:", url)
			log.Println("", "    Image:", service.Image)
			if service.Path != "" {
				log.Println("", "    Path:", service.Path)
			}
			if len(service.Ports) > 0 {
				log.Println("", fmt.Sprintf("    Ports: %v", service.Ports))
			}
			if service.Build != "" {
				log.Println("", "    Build", service.Build)
			}
			if service.Test != "" {
				log.Println("", "    Test", service.Test)
			}
			if service.Start != "" {
				log.Println("", "    Start", service.Start)
			}
		}
	}

	keen.AddEvent("bowery info", map[string]interface{}{
		"user": dev,
		"app":  state,
	})

	return 0
}
