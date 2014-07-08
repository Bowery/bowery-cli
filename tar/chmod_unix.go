// +build linux darwin

// Copyright 2013, 2014 Bowery, Inc.
package tar

import (
	"os"
)

// chmod changes the mode for a path or file desriptor.
func chmod(file *os.File, path string, mode os.FileMode) error {
	if file != nil {
		return file.Chmod(mode)
	}

	return os.Chmod(path, mode)
}
