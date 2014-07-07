// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

// handlers only take keen to log events. They return errors, so the runner reports to rollbar
type settingHandler func(keen *keen.Client, args ...string) error

var settingHandlers = map[string]settingHandler{
	"password": password,
}

func init() {
	Cmds["settings"] = &Cmd{settingsRun, "settings", "Edit your Bowery account settings.", ""}
}

func settingsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var handler settingHandler
	if len(args) > 0 {
		var ok bool
		handler, ok = settingHandlers[args[0]]
		if !ok {
			log.Println("red", "Invalid choice.")
			return 1
		}
	} else {
		log.Println("cyan", "Settings:")
		for name, _ := range settingHandlers {
			log.Println("", "\t"+name)
		}
		return 0
	}

	// if handler returns error, report it
	if err := handler(keen); err != nil {
		rollbar.Report(err)
		return 1
	}

	return 0
}

func password(keen *keen.Client, args ...string) error {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		return err
	}

	// If we have a token, we're logged in.
	if dev.Token != "" {
		log.Println("", "You're logged in as",
			strings.Split(dev.Developer.Name, " ")[0]+". Please do `bowery logout` if you wish to reset your password")
		return nil
	}

	ok, err := prompt.Ask("would you like to request a password reset")
	if !ok {
		return nil
	}

	email, err := prompt.Basic("email", true)
	if err != nil {
		return err
	}

	if err = requests.ResetPassword(email); err != nil {
		return err
	}

	log.Println("yellow", "Thank you. Check your email for a link to reset your password.")

	keen.AddEvent("broome password reset", map[string]*db.Developer{"user": dev})

	return nil
}
