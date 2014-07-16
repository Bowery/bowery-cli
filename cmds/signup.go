// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"net/mail"
	"os"
	"strings"

	"github.com/Bowery/bowery/broome"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
)

func init() {
	Cmds["signup"] = &Cmd{
		Run:   signupRun,
		Usage: "signup",
		Short: "Sign up for Bowery.",
	}
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

	developer, err := broome.CreateDeveloper(name, email, pass)
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

	if os.Getenv("ENV") == "production" && !strings.Contains(email, "@bowery.io") {
		keen.AddEvent("bowery signup", map[string]*db.Developer{"user": dev})
	}

	log.Println("magenta", "Welcome", developer.Name+"!", "To get started run",
		"`bowery connect`\nwithin your applications directory.")
	log.Println("", "")
	log.Println("magenta", "You can also check out our example apps:")
	log.Println("", "")
	log.Println("magenta", "  Node.js  https://github.com/Bowery/node-example")
	log.Println("magenta", "  Golang   https://github.com/Bowery/go-example")
	log.Println("magenta", "  Erlang   https://github.com/Bowery/erlang-example")
	log.Println("magenta", "  Spring   https://github.com/Bowery/spring-example")
	return 0
}
