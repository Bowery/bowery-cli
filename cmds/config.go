// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/prompt"
	"github.com/Bowery/cli/rollbar"
)

func init() {
	cmd := &Cmd{configRun, "config <key> [value]", "Set custom configuration options.", ""}
	cmd.Description = "Sets custom configuration options for connecting to Bowery." +
		"\n\nCurrent config options are:" +
		"\n  host  - The host bowery is running on" +
		"\n  redis - The host for a Redis connection"

	Cmds["config"] = cmd
}

func configRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var err error

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery", Cmds["config"].Usage, "\n\n"+Cmds["config"].Short)
		return 2
	}

	value := ""
	key := args[0]
	if key != "host" && key != "redis" {
		rollbar.Report(errors.ErrInvalidConfigKey)
		return 1
	}

	if len(args) < 2 {
		value, err = prompt.Basic("Value", false)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	} else {
		value = args[1]
	}

	dev, err := db.GetDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	if dev.Config == nil {
		dev.Config = make(map[string]string)
	}

	if value == "" {
		delete(dev.Config, key)
	} else {
		dev.Config[key] = value
	}

	err = dev.Save()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery config", map[string]string{})
	return 0
}
