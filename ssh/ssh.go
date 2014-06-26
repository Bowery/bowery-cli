// Copyright 2013-2014 Bowery, Inc.
// Package ssh implement routines to start a ssh shell session easily.
package ssh

import (
	"io"
	"os"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/Bowery/SkyLab/cli/errors"
	"github.com/Bowery/SkyLab/cli/log"
	"github.com/Bowery/SkyLab/cli/prompt"
	"github.com/Bowery/SkyLab/cli/schemas"
)

// Shell opens a shell connection on the servives ssh address.
func Shell(app *schemas.Application, service *schemas.Service) error {
	// Make sure we're in raw mode.
	termState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		if prompt.IsNotTerminal(err) {
			return errors.ErrIORedirection
		}

		return errors.NewStackError(err)
	}
	defer terminal.Restore(int(os.Stdin.Fd()), termState)

	// Get terminal size.
	cols, rows, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		if prompt.IsNotTerminal(err) {
			return errors.ErrIORedirection
		}

		return errors.NewStackError(err)
	}

	// Open an SSH connection to the address.
	config := &ssh.ClientConfig{User: "root", Auth: []ssh.AuthMethod{
		ssh.Password("password"),
	}}
	client, err := ssh.Dial("tcp", service.SSHAddr, config)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer client.Close()

	// Start a session on the client.
	session, err := client.NewSession()
	if err != nil {
		return errors.NewStackError(err)
	}
	defer session.Close()
	session.Stdout = prompt.NewAnsiWriter(os.Stdout)
	session.Stderr = prompt.NewAnsiWriter(os.Stderr)

	// Create a stdin pipe copying os.Stdin to it.
	stdin, err := session.StdinPipe()
	if err != nil {
		return errors.NewStackError(err)
	}
	defer stdin.Close()

	go func() {
		io.Copy(stdin, prompt.NewAnsiReader(os.Stdin))
	}()

	log.Println("magenta", "Welcome to Bowery Services.")
	log.Println("magenta", "---------------------------------------------")
	log.Println("magenta", "Name:", service.Name)
	log.Println("magenta", "Application:", app.ID)
	log.Println("magenta", "Time:", time.Now())
	log.Println("magenta", "---------------------------------------------")

	// Start a shell session.
	termModes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", rows, cols, termModes)
	if err == nil {
		err = session.Shell()
	}
	if err != nil {
		return errors.NewStackError(err)
	}

	// Wait for the session.
	err = session.Wait()
	if err != nil && err != io.EOF {
		// Ignore the error if it's an ExitError with an empty message,
		// this occurs when you do CTRL+c and then run exit cmd which isn't an
		// actual error.
		waitMsg, ok := err.(*ssh.ExitError)
		if ok && waitMsg.Msg() == "" {
			return nil
		}

		return errors.NewStackError(err)
	}

	return nil
}
