// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/SkyLab/cli/keen"
	"github.com/Bowery/SkyLab/cli/rollbar"
	"github.com/Bowery/SkyLab/cli/sys"
)

func init() {
	Cmds["support"] = &Cmd{supportRun, "support", "Create a support email.", ""}
}

func supportRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	keen.AddEvent("support command", make(map[string]string))

	err := sys.OpenPath("mailto:support@bowery.io")
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}
