// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/prompt"
	"github.com/Bowery/cli/requests"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/schemas"
)

func init() {
	Cmds["save"] = &Cmd{saveRun, "save <name>", "Save a service.", ""}
}

func saveRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var service *schemas.Service
	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["save"].Usage, "\n\n"+Cmds["save"].Short)
		return 2
	}

	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
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
	for i, v := range state.App.Services {
		services[i] = v.Name
		if args[0] == v.Name {
			service = v
			break
		}
	}

	// Handle no service found.
	if service == nil {
		log.Fprintln(os.Stderr, "red", errors.ErrInvalidService, args[0])
		log.Println("yellow", "Valid services:", strings.Join(services, ", "))
		return 1
	}
	log.Debug("Found service", service.Name, "public addr:", service.PublicAddr)

	// Get image name
	log.Println("yellow", "What would you like to call this image?")
	imageName, err := prompt.Basic("Image Name", true)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Collected Image Name", imageName)

	imageDesc, err := prompt.Basic("Description", true)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Collected Description", imageDesc)

	log.Println("yellow", "A new image is being created and saved to our registry...")
	log.Println("yellow", "This may take a couple minutes.")

	err = requests.SaveService(state, dev, service.Name, service.PublicAddr, imageName, imageDesc)
	if err != nil {
		errmsg := err.Error()
		if errmsg == imageName+" is an invalid service name" ||
			errmsg == "Image already exists" {
			log.Println("yellow", err)
		} else {
			rollbar.Report(err)
		}

		return 1
	}

	log.Println("yellow", imageName+" successfully created.")

	keen.AddEvent("bowery save", map[string]string{
		"serviceName": service.Name,
		"imageName":   imageName,
		"appId":       state.App.ID,
	})

	return 0
}
