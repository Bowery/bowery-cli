// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/keen"
	"github.com/Bowery/cli/rollbar"
)

func init() {
	Cmds["clean"] = &Cmd{cleanRun, "clean", "Reset your apps environment.", ""}
}

func cleanRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	db.GetState() // Do this to ensure we're at the root of the app.

	contents, _ := ioutil.ReadFile(filepath.Join(".bowery", "state"))
	if len(contents) > 0 {
		keen.AddEvent("bowery clean", map[string]string{
			"contents": string(contents),
		})
	}

	err := os.RemoveAll(".bowery")
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}
