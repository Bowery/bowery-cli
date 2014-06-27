// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"flag"
	"fmt"
	"os"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/prompt"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["remove"] = &Cmd{removeRun, "remove <names>", "Remove services from your application.", ""}
}

func removeRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	force := flag.Lookup("force").Value.String()

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

		if force != "true" {
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
