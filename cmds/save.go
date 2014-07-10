// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/schemas"
)

func init() {
	Cmds["save"] = &Cmd{
		Run:   saveRun,
		Usage: "save <name>",
		Short: "Save a service.",
	}
}

func saveRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["save"].Usage, "\n\n"+Cmds["save"].Short)
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

	err = api.SaveService(state, dev, service.Name, service.PublicAddr, imageName, imageDesc)
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
