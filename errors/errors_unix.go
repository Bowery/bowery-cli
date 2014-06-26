// +build linux darwin

// Copyright 2013-2014 Bowery, Inc.
package errors

import (
	"errors"
)

var (
	ErrUpdatePerm = errors.New("Unable to update, you need sudo privileges. Error Code: 21")
)
