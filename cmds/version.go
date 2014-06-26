// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Bowery/SkyLab/cli/keen"
	"github.com/Bowery/SkyLab/cli/log"
	"github.com/Bowery/SkyLab/cli/rollbar"
	"github.com/Bowery/SkyLab/cli/version"
)

func init() {
	Cmds["version"] = &Cmd{versionRun, "version", "Display the version of bowery.", ""}
}

func versionRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	keen.AddEvent("cli get version", map[string]string{"installed": version.Version})

	log.Println("", version.Version)
	return 0
}

// Compares the current version of the cli to the latest.
// If the current version is lower than the latest, return false
// otherwise return true.
func VersionOutOfDate(current, latest string) bool {
	v1 := parseVersion(current, 3)
	v2 := parseVersion(latest, 3)

	if v1 < v2 {
		return true
	}

	return false
}

// Convert "a.b.c" version to int64
func parseVersion(s string, width int) int64 {
	strList := strings.Split(s, ".")
	format := fmt.Sprintf("%%s%%0%ds", width)
	v := ""
	for _, value := range strList {
		v = fmt.Sprintf(format, v, value)
	}

	result, _ := strconv.ParseInt(v, 10, 64)
	return result
}
