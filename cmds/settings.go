// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"fmt"
	"net/mail"
	"os"
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
	Cmds["settings"] = &Cmd{
		Run:   settingsRun,
		Usage: "settings <setting>",
		Short: "Edit your Bowery account settings.",
	}
}

func settingsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	var handler settingHandler
	if len(args) > 0 {
		var ok bool
		handler, ok = settingHandlers[args[0]]
		if !ok {
			log.Fprintln(os.Stderr, "red", "Invalid choice.")
			return 1
		}
	} else {
		fmt.Fprintln(os.Stderr,
			"Usage: bowery", Cmds["settings"].Usage, "\n\n"+Cmds["settings"].Short+"\n")
		fmt.Fprintln(os.Stderr, "Settings:")
		for name, _ := range settingHandlers {
			fmt.Fprintln(os.Stderr, " ", name)
		}
		return 2
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

	if err == nil {
		log.Println("yellow", "You're logged in as", strings.Split(dev.Developer.Name, " ")[0]+
			". Please do `bowery logout` before trying to reset your password.")
		return nil
	}

	ok, err := prompt.Ask("Would you like to request a password reset")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	valid := false
	email := ""
	for !valid {
		email, err = prompt.Basic("Email", true)
		if err != nil {
			return err
		}
		_, err = mail.ParseAddress(email)
		if err != nil {
			log.Println("yellow", "Try again! Valid email address required.")
		} else {
			valid = true
		}
	}

	if err = requests.ResetPassword(email); err != nil {
		return err
	}

	log.Println("", "Thank you. Check your email for a link to reset your password.")
	keen.AddEvent("broome password reset", map[string]string{"email": email})
	return nil
}
