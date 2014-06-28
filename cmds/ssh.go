// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/cli/ssh"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/schemas"
)

func init() {
	Cmds["ssh"] = &Cmd{sshRun, "ssh <name>", "Connect to a service via ssh.", ""}
}

func sshRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var service *schemas.Service
	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery "+Cmds["ssh"].Usage, "\n\n"+Cmds["ssh"].Short)
		return 2 // --help uses 2.
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
		}
	}

	// Handle no service found.
	if service == nil {
		log.Fprintln(os.Stderr, "red", errors.ErrInvalidService, args[0])
		log.Println("yellow", "Valid services:", strings.Join(services, ", "))
		return 1
	} else {
		log.Debug("Found service", service.Name, "ssh addr:", service.SSHAddr)
	}

	err = ssh.Shell(state.App, service)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}
