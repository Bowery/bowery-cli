// Copyright 2013-2014 Bowery, Inc.
// Package errors provides error messages and routines to provide
// error types with stack information.
package errors

import (
	"errors"
	pkgerr "github.com/Bowery/gopackages/errors"
)

// Standard errors that may occur.
var (
	ErrNoDeveloper      = errors.New("No developer found. Run `bowery connect`. Error Code: 0")
	ErrDeveloperExists  = errors.New("Developer with email already exists. Error Code: 1")
	ErrContactSupport   = errors.New("Contact support with the message below. Run `bowery support`.")
	ErrNotConnected     = errors.New("You're not connected. Run `bowery connect`. Error Code: 2")
	ErrEmpty            = errors.New("may not be empty.")
	ErrInvalidCommand   = errors.New("Invalid command. Error Code: 3")
	ErrInvalidService   = errors.New("Invalid service. Error Code: 4")
	ErrInvalidLogin     = errors.New("Invalid email/password. Error Code: 5")
	ErrTooManyLogins    = errors.New("Too many incorrect login attempts. Error Code: 6")
	ErrNoServices       = errors.New("You haven't registered any services. Run `bowery add` to create services. Error Code: 7")
	ErrNoServicePaths   = errors.New("No services to sync, add a path to at least one service to sync file changes. Error Code: 8")
	ErrCantConnect      = errors.New("Looks like there is a problem connecting to Bowery. Error Code: 9")
	ErrFailedRestart    = errors.New("Error restarting your service. Error Code: 10")
	ErrMismatchPass     = errors.New("Your passwords do not match. Error Code: 11")
	ErrOutOfDate        = errors.New("Bowery is out of date. Run `bowery update`. Error Code: 12")
	ErrVersionDownload  = errors.New("An error occured getting the bowery version. Error Code: 13")
	ErrSyncFailed       = errors.New("Cannot sync the service path. Run `bowery restart` if this problem persists. Error Code: 14")
	ErrImageExists      = errors.New("Image already exists. Try a different name. Error Code: 15")
	ErrNoImageFound     = errors.New("No image found. Error Code: 16")
	ErrCTRLC            = errors.New("Interrupted (CTRL+C)")
	ErrIORedirection    = errors.New("Input/output may not be redirected, must be ran with console. Error Code: 17")
	ErrInvalidConfigKey = errors.New("Invalid config key. View `bowery help config` Error Code: 22")
	ErrInvalidToken     = errors.New("Invalid token, log out and back in to update credentials. Error Code: 23")
	ErrOverCapacity     = errors.New("Bowery is momentarily over capacity. Please try again. Error Code: 24")
	ErrContainerConnect = errors.New("Unable to connect to container. Run `bowery restart` if this problem persists. Error Code: 25")
	ErrResetRequest     = errors.New("Unable to reset your password. Please Try again. Error Code: 26")
	ErrInvalidEmail     = errors.New("Email does not match an existing user.")
)

// Error templates to be used with Newf.
const (
	ErrPathNotFoundTmpl = "The path for %s(%s) does not exist. Create it to continue. Error Code: 18"
	ErrPathNotDirTmpl   = "The path for %s(%s) is not a directory. Error Code: 27"
	ErrSyncTmpl         = "(%s): %s"
	ErrLoginRetryTmpl   = "%s Try again."
	ErrInvalidJSONTmpl  = "Invalid JSON in file %s. Error Code: 19"
	ErrInvalidPortTmpl  = "%s is an invalid port. Try again. Error Code: 20"
	ErrErrorsRange      = "Valid range: 0 - %d"
)

// Error function wrappers.
var (
	New           = pkgerr.New
	Newf          = pkgerr.Newf
	IsStackError  = pkgerr.IsStackError
	NewStackError = pkgerr.NewStackError
)
