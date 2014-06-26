// Copyright 2013-2014 Bowery, Inc.
package db

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/schemas"
)

// State describes a bowery applications saved state.
type State struct {
	Token  string               `json:"token,omitempty"`
	App    *schemas.Application `json:"app"`
	Config map[string]*Service  `json:"config"`
	Path   string               `json:"-"`
}

// GetState retrieves the applications state.
func GetState() (*State, error) {
	state := new(State)
	state.Path = filepath.Join(".bowery", "state")

	err := jumpToParent(state.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = load(state, state.Path)
	if err == io.EOF || os.IsNotExist(err) {
		err = errors.ErrNotConnected
	}

	return state, err
}

// Save writes the state to the state path.
func (state *State) Save() error {
	state.Token = "" // Token shouldn't be saved in state, used for requests.

	return save(state, state.Path)
}
