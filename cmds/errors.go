// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/rollbar"
)

func init() {
	Cmds["errors"] = &Cmd{errorsRun, "errors [id]", "Get information on an error.", ""}
}

func errorsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	if len(args) == 0 {
		log.Println("cyan", "All Errors. Run `bowery errors <code>` for more info.\n")
		tabWriter := tabwriter.NewWriter(os.Stderr, 0, 0, 1, ' ', 0)
		for i, err := range errors.GetAll() {
			desc := strings.Split(err.Title, "Error Code")[0]
			fmt.Fprintln(tabWriter, i, "\t"+desc)
		}
		tabWriter.Flush()
		return 0
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		log.Println("yellow", "Input must be an integer.")
		return 1
	}

	info, err := errors.Get(index)
	if err != nil {
		log.Println("yellow", err)
		return 1
	}

	desc := strings.Split(info.Title, "Error Code")[0]

	log.Print("cyan", "Code: ")
	log.Println("", info.Code)
	log.Print("cyan", "Title: ")
	log.Println("", desc)
	log.Println("cyan", "Description:")
	log.Println("", info.Description)

	return 0
}
