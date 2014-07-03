// Copyright 2013-2014 Bowery, Inc.
// Package sys implements routines that would otherwise require
// extra code to make the same across operating systems.
package sys

import (
	"os/exec"

	"github.com/Bowery/bowery/errors"
)

// OpenPath opens a path using the systems preferred application.
func OpenPath(args ...string) error {
	cmd := exec.Command(openProg[0], append(openProg[1:], args...)...)
	err := cmd.Run()
	_, ok := err.(*exec.ExitError)
	if err != nil && !ok {
		return errors.NewStackError(err)
	}

	return nil
}
