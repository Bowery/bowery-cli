// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"net/mail"
	"os"
	"strings"

	"github.com/Bowery/SkyLab/cli/db"
	"github.com/Bowery/SkyLab/cli/errors"
	"github.com/Bowery/SkyLab/cli/keen"
	"github.com/Bowery/SkyLab/cli/log"
	"github.com/Bowery/SkyLab/cli/prompt"
	"github.com/Bowery/SkyLab/cli/requests"
	"github.com/Bowery/SkyLab/cli/rollbar"
)

func init() {
	Cmds["login"] = &Cmd{loginRun, "login", "Login in to your Bowery account.", ""}
}

func loginRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If we have a token, we're logged in.
	if dev.Token != "" {
		log.Println("", "You're logged in as",
			strings.Split(dev.Developer.Name, " ")[0]+".")
		return 0
	}

	err = getToken(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	err = getDeveloper(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery login", map[string]*db.Developer{"user": dev})

	log.Println("magenta", "Hey there", strings.Split(dev.Developer.Name, " ")[0],
		"you're logged in now.")
	return 0
}

func getToken(dev *db.Developer) error {
	var err error
	i := 0
	token := ""

	// Get email and password up to 5 times, then report the error.
	for token == "" && i < 5 {
		validEmail := false
		email := ""
		pass := ""

		for !validEmail {
			email, err = prompt.Basic("Email", true)
			if err != nil {
				return err
			}

			_, err = mail.ParseAddress(email)
			if err == nil {
				validEmail = true
			} else {
				log.Println("yellow", "Try again! Valid email address required.")
			}
		}

		pass, err = prompt.Password("Password")
		if err != nil {
			return err
		}
		log.Debug("Collected email", email, "pass", pass)

		token, err = requests.GetTokenByLogin(email, pass)
		if err != nil {
			if i < 4 {
				log.Fprintln(os.Stderr, "red", errors.Newf(errors.ErrLoginRetryTmpl, err))
			}
			i++
		}
	}

	if err != nil {
		if err == errors.ErrInvalidLogin {
			err = errors.ErrTooManyLogins
		}

		return err
	}

	log.Debug("Got token", token)
	dev.Token = token
	return dev.Save()
}

func getDeveloper(dev *db.Developer) error {
	// Get the developer from the devs token.
	developer, err := requests.GetDeveloper(dev.Token)
	if err != nil {
		return err
	}

	// Save the developer.
	log.Debug("Found developer", developer)
	dev.Developer = developer
	return dev.Save()
}
