// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/bowery/sys"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["invite"] = &Cmd{inviteRun, "invite", "Invite a friend to use Bowery.", ""}
}

func inviteRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	keen.AddEvent("bowery invite", make(map[string]string))

	err := sys.OpenPath("mailto:?subject=You%20should%20try%20out%20Bowery&body=Hey!%20I've%20been%20using%20Bowery%20and%20thought%20you'd%20like%20to%20test%20it%20out%3A%0A%0Ahttp%3A%2F%2Fbowery.io")
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}
