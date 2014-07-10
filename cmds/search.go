// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["search"] = &Cmd{
		Run:   searchRun,
		Usage: "search [image]",
		Short: "Search by name for available images.",
	}
}

func searchRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	if len(args) <= 0 {
		args = append(args, "")
	}

	for _, name := range args {
		// If empty send "." which returns all images.
		if name == "" {
			name = "."
		}

		images, err := api.SearchImages(name)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		if len(images) <= 0 {
			log.Println("", "No Result for '"+name+"' were found.")
		} else {
			if name == "." {
				log.Println("", "All images:")
			} else {
				log.Println("", "Search Results for '"+name+"':")
			}
		}

		for _, image := range images {
			if image.Description == "" {
				log.Println("", "  -", image.Name)
			} else {
				log.Println("", "  -", image.Name, "-", image.Description)
			}
		}
	}
	return 0
}
