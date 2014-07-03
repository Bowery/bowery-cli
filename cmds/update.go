// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"io"
	"os"
	"path/filepath"

	"bitbucket.org/kardianos/osext"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/bowery/version"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["update"] = &Cmd{updateRun, "update", "Update Bowery.", ""}
}

func updateRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	keen.AddEvent("cli update", map[string]string{"installed": version.Version})

	ver, err := requests.GetVersion()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	if ver == version.Version {
		log.Println("", "Bowery is up to date.")
		return 0
	}
	log.Println("yellow", "Bowery is out of date. Updating to", ver, "now...")

	newVer, err := requests.DownloadNewVersion(ver)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	exec, err := osext.Executable()
	if err != nil {
		rollbar.Report(errors.NewStackError(err))
		return 1
	}
	tempExec := filepath.Join(filepath.Dir(exec), ".old_bowery"+filepath.Ext(exec))

	// Open exec, should fail if execing somewhere else.
	file, err := os.Open(exec)
	if err != nil {
		rollbar.Report(errors.ErrUpdatePerm)
		return 1
	}
	file.Close()

	// Create the temp exec file to test io permissions.
	file, err = os.Create(tempExec)
	if err != nil {
		rollbar.Report(errors.ErrUpdatePerm)
		return 1
	}
	file.Close()

	// Remove it which also removes any previous executables.
	err = os.RemoveAll(tempExec)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	// Move the exec to a temp file so we can write the new one.
	err = os.Rename(exec, tempExec)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	file, err = os.OpenFile(exec, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	defer file.Close()

	_, err = io.Copy(file, newVer)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	log.Println("magenta", "Updated bowery to version", ver+".")
	return 0
}
