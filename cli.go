// Copyright 2013-2014 Bowery, Inc.
package main

import (
	"flag"
	"os"
	"path/filepath"

	"bitbucket.org/kardianos/osext"
	. "github.com/Bowery/cli/cmds"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/rollbar"
)

var (
	env = os.Getenv("ENV")
)

func init() {
	flag.Bool("force", false, "Force pull.")
}

func main() {
	if env == "" {
		env = "production"
	}

	// Remove any previous executables, ignore errors as it isn't essential.
	exec, _ := osext.Executable()
	if exec != "" {
		os.Remove(filepath.Join(filepath.Dir(exec), ".old_bowery"+filepath.Ext(exec)))
	}

	// Set up error and analytics reporting.
	keen := &keen.Client{
		WriteKey:  "8bbe0d9425a22a6c31e6da9ae3012c738ee21000b533c351a419bb0e3d08431456359d1bea654a39c2065df0b1df997ecde7e3cf49a9be0cd44341b15c1ff5523f13d26d8060373390f47bcc6a33b80e69e2b2c1101cde4ddb3d20b16a53a439a98043919e809c09c30e4856dedc963f",
		ProjectID: "52c08d6736bf5a4a4b000005",
	}
	rollbar := &rollbar.Client{
		Token: "a7c4e78074034f04b1882af596657295",
		Env:   env,
	}

	// Parse flags and get arguments.
	flag.Usage = func() {
		Cmds["help"].Run(keen, rollbar)
	}

	flag.Parse()
	args := flag.Args()
	command := "help"

	if len(args) >= 1 {
		command = args[0]
		args = args[1:]
	}

	// Run command, and handle invalid commands.
	cmd, ok := Cmds[command]
	if !ok {
		keen.AddEvent("invalid command", map[string]string{"command": command})

		log.Fprintln(os.Stderr, "red", errors.ErrInvalidCommand, command)
		os.Exit(Cmds["help"].Run(keen, rollbar))
	}

	os.Exit(cmd.Run(keen, rollbar, args...))
}
