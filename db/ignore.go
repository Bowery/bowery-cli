// Copyright 2013-2014 Bowery, Inc.
package db

import (
	"bufio"
	"os"
	"path/filepath"
)

// GetIgnores retrieves a list of files to ignore in a given directory.
func GetIgnores(dir string) ([]string, error) {
	ignores := []string{".bowery", ".hg", ".git", ".svn", ".bzr"}
	matches := make([]string, 0)

	file, err := os.Open(filepath.Join(dir, ".boweryignore"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if file != nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			ignores = append(ignores, scanner.Text())
		}

		err := scanner.Err()
		if err != nil {
			return nil, err
		}
	}

	for _, ignore := range ignores {
		ignoreMatches, err := filepath.Glob(filepath.Join(dir, ignore))
		if err != nil {
			return nil, err
		}

		matches = append(matches, ignoreMatches...)
	}

	return matches, nil
}
