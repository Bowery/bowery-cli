// Copyright 2013-2014 Bowery, Inc.
// Package errors provides error messages and routines to provide
// error types with stack information.
package errors

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
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
	ErrInvalidEmail = errors.New("Email does not match an existing user.")
)

// Error templates to be used with Newf.
const (
	ErrPathNotFoundTmpl = "The path for %s(%s) does not exist. Create it to continue. Error Code: 18"
	ErrSyncTmpl         = "(%s): %s"
	ErrLoginRetryTmpl   = "%s Try again."
	ErrInvalidJSONTmpl  = "Invalid JSON in file %s. Error Code: 19"
	ErrInvalidPortTmpl  = "%s is an invalid port. Try again. Error Code: 20"
	ErrErrorsRange      = "Valid range: 0 - %d"
)

// New creates new error, this solves issue of name collision with
// errors pkg.
func New(args ...interface{}) error {
	return errors.New(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Newf creates a new error, from an existing error template.
func Newf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}

// StackError is an error with stack information.
type StackError struct {
	Err   error
	Trace *Trace
}

// IsStackError returns the error as a StackError if it's a StackError, nil
// otherwise.
func IsStackError(err error) *StackError {
	se, ok := err.(*StackError)
	if ok {
		return se
	}

	return nil
}

// NewStackError creates a stack error including the stack.
func NewStackError(err error) error {
	se := &StackError{
		Err: err,
		Trace: &Trace{
			Frames:    make([]*frame, 0),
			Exception: &exception{Message: err.Error(), Class: errClass(err)},
		},
	}

	// Get stack frames excluding the current one.
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			// Couldn't get another frame, so we're finished.
			break
		}

		f := &frame{File: file, Line: line, Method: routineName(pc)}
		se.Trace.Frames = append(se.Trace.Frames, f)
	}

	return se
}

func (se *StackError) Error() string {
	return se.Err.Error()
}

// Stack prints the stack trace in a readable format.
func (se *StackError) Stack() string {
	stack := ""

	for i, frame := range se.Trace.Frames {
		stack += strconv.Itoa(i+1) + ": File \"" + frame.File + "\" line "
		stack += strconv.Itoa(frame.Line) + " in " + frame.Method + "\n"
	}
	stack += se.Trace.Exception.Class + ": " + se.Trace.Exception.Message

	return stack
}

// Trace contains the stack frames, and the exception information.
type Trace struct {
	Frames    []*frame   `json:"frames"`
	Exception *exception `json:"exception"`
}

// exception contains the error message and it's class origin.
type exception struct {
	Class   string `json:"class"`
	Message string `json:"message"`
}

// frame contains line, file and method info for a stack frame.
type frame struct {
	File   string `json:"filename"`
	Line   int    `json:"lineno"`
	Method string `json:"method"`
}

// errClass retrieves the string representation for the errors type.
func errClass(err error) string {
	class := strings.TrimPrefix(reflect.TypeOf(err).String(), "*")
	if class == "" {
		class = "panic"
	}

	return class
}

// routineName returns the routines name for a given program counter.
func routineName(pc uintptr) string {
	fc := runtime.FuncForPC(pc)
	if fc == nil {
		return "???"
	}

	return fc.Name() // Includes the package info.
}
