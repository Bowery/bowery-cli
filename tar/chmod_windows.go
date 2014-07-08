// Copyright 2013, 2014 Bowery, Inc.
package tar

import (
	"os"
	"syscall"
)

// chmod changes the mode for a path or file descriptor. It catches any
// unsupported errors.
func chmod(file *os.File, path string, mode os.FileMode) error {
	var err error
	if file != nil {
		err = file.Chmod(mode)
	} else {
		err = os.Chmod(path, mode)
	}

	pe, ok := err.(*os.PathError)
	if ok && pe.Err == syscall.EWINDOWS {
		err = nil
	}
	return err
}
