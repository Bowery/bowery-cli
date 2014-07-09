// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["remove"] = &Cmd{
		Run:   removeRun,
		Usage: "remove <names>",
		Short: "Remove services from your application.",
	}
}

func removeRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := Cmds["remove"].Force

	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["remove"].Usage, "\n\n"+Cmds["remove"].Short)
		return 2 // --help uses 2.
	}

	services, err := db.GetServices()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	for _, name := range args {
		_, ok := services.Data[name]
		if !ok {
			log.Println("yellow", "Service", name, "doesn't exist, skipping.")
			continue
		}

		if !force {
			ok, err = prompt.Ask("Are you sure you want to remove " + name)
			if err != nil {
				rollbar.Report(err)
				return 1
			}
			if !ok {
				log.Println("yellow", "Skipping", name)
				continue
			}
		}

		delete(services.Data, name)
		log.Println("", "Removed service", name)
	}

	err = services.Save()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery remove", map[string]interface{}{
		"toRemove": args,
		"existing": services.Data,
	})
	return 0
}
