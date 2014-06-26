// Copyright 2013-2014 Bowery, Inc.
// Package db implements routines to retrieve and save data
// about apps and devs on the FS.
package db

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Bowery/cli/errors"
)

// jumpToParent looks for a path up 5 directories and changes to the directory
// the path is found in.
func jumpToParent(path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	prevCmd := ""

	for i := 0; i < 5; i++ {
		_, err = os.Lstat(filepath.Join(cwd, path))
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// Not found.
		if err != nil {
			prevCmd = cwd
			cwd = filepath.Dir(cwd)

			// This occurs if we've reached the root.
			if prevCmd == cwd {
				return os.ErrNotExist
			}
			continue
		}

		return os.Chdir(cwd) // Found.
	}

	return os.ErrNotExist
}

// load decodes a json file into the given data.
func load(data interface{}, path string) error {
	// Ensure parent exists.
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
	if err != nil {
		return err
	}

	// Open for decoding.
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the contents and into the given data.
	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return errors.Newf(errors.ErrInvalidJSONTmpl, path)
	}

	return nil
}

// save encodes the given data to a json file, which is created if needed.
func save(data interface{}, path string) error {
	// Ensure parent exists.
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
	if err != nil {
		return err
	}

	// Open for writing.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marshal data with indents, for human readability.
	dat, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(dat)

	_, err = io.Copy(file, buf)
	return err
}
