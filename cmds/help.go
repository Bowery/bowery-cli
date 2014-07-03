// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["help"] = &Cmd{helpRun, "help [command]", "Display usage for commands.", ""}
}

func helpRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	if len(args) > 0 {
		cmd, ok := Cmds[args[0]]
		if !ok {
			log.Fprintln(os.Stderr, "red", errors.ErrInvalidCommand, args[0])
			return 1
		}

		fmt.Fprintln(os.Stderr, "Usage: bowery", cmd.Usage+"\n")
		if cmd.Description == "" {
			cmd.Description = cmd.Short
		}

		fmt.Fprintln(os.Stderr, cmd.Description)
		return 2 // --help uses 2.
	}

	// Ensure output is correctly aligned.
	tabWriter := tabwriter.NewWriter(os.Stderr, 0, 0, 8, ' ', 0)
	fmt.Fprintln(os.Stderr, "Usage: bowery [option] <command> [args]\n")
	fmt.Fprintln(os.Stderr, "Options:\n  --force Force actions instead of asking.\n")
	fmt.Fprintln(os.Stderr, "Commands:")

	for _, cmd := range Cmds {
		// \t is used to separate columns.
		fmt.Fprintln(tabWriter, "  "+cmd.Usage+"\t"+cmd.Short)
	}
	tabWriter.Flush()
	return 2 // --help uses 2.
}
