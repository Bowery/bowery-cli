// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/requests"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/schemas"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["restart"] = &Cmd{restartRun, "restart <name>", "Restart a service.", ""}
}

func restartRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var service *schemas.Service

	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["restart"].Usage, "\n\n"+Cmds["restart"].Short)
		return 2
	}

	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	if dev.Token == "" {
		log.Println("yellow", "Oops! You must be logged in.")

		err = getToken(dev)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	err = getDeveloper(dev)
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

	newService, err := requests.RestartService(service.DockerID, dev.Token)
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
