// Copyright 2013-2014 Bowery, Inc.
package errors

import (
	"errors"
)

var (
	ErrUpdatePerm = errors.New("Unable to update, run cmd as administrator. Error Code: 21")
)
