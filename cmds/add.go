// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"strconv"
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["add"] = &Cmd{addRun, "add [names]", "Add services to your application.", ""}
}

func addRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	includedName := true
	services, err := db.GetServices()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	// If no given arguments we want to add a single item.
	if len(args) <= 0 {
		includedName = false
		args = append(args, "")

		log.Println("cyan bold", "New Service Wizard\n")
		log.Println("magenta", "The basis for a service is it's image. Bowery provides a number of")
		log.Println("magenta", "\"images\" all of which you can find at directory.bowery.io. You")
		log.Println("magenta", "can also run `bowery search` to look them up from the command line.\n")
	}

	for _, name := range args {
		// Get name if none were given.
		if name == "" {
			log.Println("magenta", "What would you like to call this service? (e.g. api, db, cache, imageprocessor, etc)")
			name, err = prompt.Basic("Name", true)
			if err != nil {
				rollbar.Report(err)
				return 1
			}
		}

		// Normalize name
		name = strings.ToLower(name)
		for _, c := range "\f\n\r\t\v\u00A0\u2028\u2029" {
			name = strings.Replace(name, string(c), "-", -1)
		}

		// If name already exists, prompt for overwrite.
		_, ok := services.Data[name]
		if ok {
			ok, err = prompt.Ask("Do you want to overwrite " + name)
			if err != nil {
				rollbar.Report(err)
				return 1
			}
			if !ok {
				log.Println("yellow", "Skipping", name)
				continue
			}
		}

		if includedName && len(args) > 1 {
			log.Println("magenta", "Creating", name)
		}

		// Ask for image name
		validImage := false
		image := ""
		if !includedName {
			log.Println("magenta", "What image would you like to use? (e.g. php, mongodb, ruby, node, etc.)")
		}
		for validImage == false {
			image, err = prompt.Basic("Image", false)
			if err != nil {
				rollbar.Report(err)
				return 1
			}

			if image != "base" {
				err := requests.FindImage(image)
				if err != nil && err != errors.ErrNoImageFound {
					rollbar.Report(err)
					return 1
				}
			}

			if err == errors.ErrNoImageFound {
				ok, err = prompt.Ask("Invalid image type. Would you like to use the base image")
				if err != nil {
					rollbar.Report(err)
					return 1
				}

				if !ok {
					log.Println("yellow", "Try another image. Search for them using `bowery search`.")
					continue
				}

				image = "base"
			}

			validImage = true
		}

		// Ask for path
		if !includedName {
			log.Println("magenta", "What files would you like to sync?")
		}
		path, err := prompt.Basic("Path", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		// Ask for ports
		if !includedName {
			log.Println("magenta", "What ports would you like exposed? Enter comma separated. Ports 22 and 80 are included by default, 3001 is reserved.")
		}
		ports, err := prompt.Basic("Ports", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
		var portsList []interface{}

		// Create list of int ports.
		if ports != "" {
			portsSplit := strings.Split(ports, ",")
			portsList = make([]interface{}, len(portsSplit))
			for i, port := range portsSplit {
				port = strings.Trim(port, " ")
				num, err := strconv.Atoi(port)
				if err != nil {
					rollbar.Report(errors.Newf(errors.ErrInvalidPortTmpl, port))
					return 1
				}

				portsList[i] = num
			}
		}

		// Ask for start
		if !includedName {
			log.Println("magenta", "How does the service start? (e.g. node app.js, mongod, etc.)")
		}
		start, err := prompt.Basic("Start Command", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		// Ask for build
		if !includedName {
			log.Println("magenta", "How does the service get built? (e.g. bundle install, ./configure && make, etc.)")
		}
		build, err := prompt.Basic("Build Command", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		// Ask for test
		if !includedName {
			log.Println("magenta", "How does the service get tested? (e.g. make test, npm test, etc.)")
		}
		test, err := prompt.Basic("Test Command", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		log.Debug("Adding service", "name", name, "image", image, "path", path, "ports", portsList, "start", start, "build", build, "test", test)
		services.Data[name] = &db.Service{
			Image: image,
			Path:  path,
			Ports: portsList,
			Start: start,
			Build: build,
			Test:  test,
		}
	}

	log.Debug("Saving services", services.Data)
	err = services.Save()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery add", services.Data)
	return 0
}
