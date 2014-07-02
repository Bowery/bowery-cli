// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/sys"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["invite"] = &Cmd{inviteRun, "invite", "Invite a friend to use Bowery.", ""}
}

func inviteRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	keen.AddEvent("invite command", make(map[string]string))

	err := sys.OpenPath("mailto: ?subject=You+should+try+out+Bowery&body=Hey%21+I%27ve+been+using+Bowery+and+thought+you%27d+like+to+test+it+out%3A%0A%0Ahttp%3A%2F%2Fbowery.io")
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}
