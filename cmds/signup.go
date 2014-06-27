// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"net/mail"
	"os"

	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/errors"
	"github.com/Bowery/cli/log"
	"github.com/Bowery/cli/prompt"
	"github.com/Bowery/cli/requests"
	"github.com/Bowery/cli/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["signup"] = &Cmd{signupRun, "signup", "Sign up for Bowery.", ""}
}

func signupRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	log.Println("magenta", "Welcome to Bowery! Just provide us some basic info",
		"\nabout yourself and we'll get you up and running.")
	name, err := prompt.Basic("Name", true)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Collected name", name)

	validEmail := false
	email := ""
	for !validEmail {
		email, err = prompt.Basic("Email", true)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
		_, err = mail.ParseAddress(email)
		if err == nil {
			validEmail = true
		} else {
			log.Println("yellow", "Try again! Valid email address required.")
		}
	}

	log.Debug("Collected email", email)

	pass, err := prompt.Password("Password")
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Collected password", pass)

	conf, err := prompt.Password("Confirm Password")
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Debug("Collected confirmation password", conf)

	if pass != conf {
		log.Fprintln(os.Stderr, "red", errors.ErrMismatchPass, "Try again.")
		return 1
	}

	developer, err := requests.CreateDeveloper(name, email, pass)
	if err != nil {
		if err == errors.ErrDeveloperExists {
			log.Println("yellow", err)
		} else {
			rollbar.Report(err)
		}

		return 1
	}
	log.Debug("Created developer", developer)

	dev.Token = developer.Token
	dev.Developer = developer
	err = dev.Save()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery signup", map[string]*db.Developer{"user": dev})

	log.Println("magenta", "Welcome", developer.Name+"!", "To get started run",
		"`bowery connect`\nwithin your applications directory.")
	log.Println("magenta", "\nYou can also check out the example app at https://github.com/Bowery/example\nor run `bowery help`.")
	return 0
}
