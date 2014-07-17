// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
	"github.com/Bowery/gopackages/schemas"
)

func init() {
	Cmds["info"] = &Cmd{
		Run:   infoRun,
		Usage: "info",
		Short: "Display developer info and application info.",
	}
}

func infoRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, devErr := db.GetDeveloper()
	if devErr != nil && devErr != errors.ErrNoDeveloper {
		rollbar.Report(devErr)
		return 1
	}

	state, err := db.GetState()
	if err != nil && err != errors.ErrNotConnected {
		rollbar.Report(err)
		return 1
	}

	if ENV == "development" {
		log.Println("magenta", "Developer(development mode):")
	} else {
		log.Println("magenta", "Developer:")
	}
	if devErr != nil {
		log.Println("yellow", "  Not logged in.")
	} else {
		log.Println("", "  Email:", dev.Developer.Email)
		log.Println("", "  Name:", dev.Developer.Name)
	}

	log.Println("magenta", "\nConfig:")
	if dev.Config == nil || len(dev.Config) <= 0 {
		log.Println("yellow", "  No custom configuration.")
	} else {
		for k, v := range dev.Config {
			if v != "" {
				log.Println("", "  "+strings.Title(k)+":", v)
			}
		}
	}

	log.Println("magenta", "\nApplication:")
	if err == errors.ErrNotConnected {
		log.Println("yellow", "  No connected application in this directory.")
	} else {
		if state.App.Name != "" {
			log.Println("", "  Name:", state.App.Name)
		}
		log.Println("", "  ID:", state.App.ID)

		for name, service := range state.Config {
			var apiService *schemas.Service
			for _, serv := range state.App.Services {
				if serv.Name == name {
					apiService = serv
					break
				}
			}

			url := apiService.PublicAddr
			if ENV != "development" {
				if state.App.Name != "" {
					url = "http://" + name + "." + state.App.Name + ".boweryapps.com"
				} else {
					url = "http://" + name + "." + state.App.ID + ".boweryapps.com"
				}
			}

			log.Println("", "  "+name+":")
			log.Println("", "    URL:", url)
			log.Println("", "    Image:", service.Image)
			if service.Path != "" {
				paths := strings.Split(service.Path, ":")

				log.Println("", "    Path:", paths[0])
				if len(paths) > 1 && paths[1] != "" {
					log.Println("", "    Remote Path:", paths[1])
				}
			}
			if len(service.Ports) > 0 {
				log.Println("", fmt.Sprintf("    Ports: %v", service.Ports))
			}
			if service.Build != "" {
				log.Println("", "    Build:", service.Build)
			}
			if service.Test != "" {
				log.Println("", "    Test:", service.Test)
			}
			if service.Start != "" {
				log.Println("", "    Start:", service.Start)
			}
			if len(service.Env) > 0 {
				log.Println("", "    Env:")
				for key, value := range service.Env {
					log.Println("", fmt.Sprintf("      %s: %s", key, value))
				}
			}
		}
	}

	keen.AddEvent("bowery info", map[string]interface{}{
		"user": dev,
		"app":  state,
	})
	return 0
}
