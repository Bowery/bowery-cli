// Copyright 2013-2014 Bowery, Inc.
package db

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/sys"
	"github.com/Bowery/gopackages/schemas"
)

var env = os.Getenv("ENV")

// Developer is a developer including token and api developer struct.
type Developer struct {
	Token     string             `json:"token"`
	Developer *schemas.Developer `json:"developer"`
	Config    map[string]string  `json:"config"`
	path      string
}

// GetDeveloper retrieves the developer from their user config.
func GetDeveloper() (*Developer, error) {
	dev := new(Developer)

	dev.path = ".boweryconf"
	if env == "development" {
		dev.path = ".bowerydevconf"
	}
	dev.path = filepath.Join(os.Getenv(sys.HomeVar), dev.path)

	err := load(dev, dev.path)
	if err == io.EOF || os.IsNotExist(err) {
		err = errors.ErrNoDeveloper
	}

	return dev, err
}

// Save writes the developer to their db path.
func (dev *Developer) Save() error {
	return save(dev, dev.path)
}
