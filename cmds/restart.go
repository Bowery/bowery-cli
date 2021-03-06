// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
	"github.com/Bowery/gopackages/schemas"
)

func init() {
	Cmds["restart"] = &Cmd{
		Run:   restartRun,
		Usage: "restart <name>",
		Short: "Restart a service.",
	}
}

func restartRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["restart"].Usage, "\n\n"+Cmds["restart"].Short)
		return 2
	}

	dev, err := getDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	state, err := db.GetState()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	// Create slices of service names, and find the requested service.
	var service *schemas.Service
	services := make([]string, len(state.App.Services))
	serviceIdx := -1
	for i, v := range state.App.Services {
		services[i] = v.Name
		if args[0] == v.Name {
			service = v
			serviceIdx = i
			break
		}
	}

	// Handle no service found.
	if service == nil {
		log.Fprintln(os.Stderr, "red", errors.ErrInvalidService, args[0])
		log.Println("yellow", "Valid services:", strings.Join(services, ", "))
		return 1
	}
	log.Debug("Found service", service.Name)

	newService, err := api.RestartService(service.DockerID, dev.Token)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	if newService != nil {
		state.App.Services[serviceIdx] = newService
		service = newService

		err = state.Save()
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	keen.AddEvent("bowery restart", map[string]string{
		"name":  service.Name,
		"appId": state.App.ID,
	})
	return 0
}
