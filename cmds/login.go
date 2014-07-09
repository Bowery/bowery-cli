// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"net/mail"
	"os"
	"strings"

	"github.com/Bowery/bowery/broome"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/prompt"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"
)

func init() {
	Cmds["login"] = &Cmd{
		Run:   loginRun,
		Usage: "login",
		Short: "Login to your Bowery account.",
	}
}

func loginRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If dev was found then check if token is up to date.
	if err == nil {
		ok, err := devUpToDate(dev)
		if err != nil {
			rollbar.Report(err)
			return 1
		}

		if ok {
			log.Println("", "You're logged in as",
				strings.Split(dev.Developer.Name, " ")[0]+".")
			return 0
		} else {
			log.Println("yellow", "Oops! Your login information is out of date.")
		}
	}

	err = getToken(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	err = updateDeveloper(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery login", map[string]*db.Developer{"user": dev})

	log.Println("magenta", "Hey there", strings.Split(dev.Developer.Name, " ")[0],
		"you're logged in now.")
	return 0
}

// getDeveloper retrieves the local dev and updates if out of date, or no
// dev exists.
func getDeveloper() (*db.Developer, error) {
	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		return dev, err
	}

	ok := false
	if err != nil {
		log.Println("yellow", "Oops! You must be logged in.")
	} else {
		ok, err = devUpToDate(dev)
		if err != nil {
			return dev, err
		}

		if !ok {
			log.Println("yellow", "Oops! Your login information is out of date.")
		}
	}

	if err != nil || !ok {
		err = getToken(dev)
		if err != nil {
			return dev, err
		}
	}

	return dev, updateDeveloper(dev)
}

// devUpToDate checks if a developers token is up to date.
func devUpToDate(dev *db.Developer) (bool, error) {
	remoteDev, err := broome.GetDeveloper(dev.Token)
	if err != nil && err != errors.ErrInvalidToken {
		return false, err
	}

	if remoteDev != nil && dev.Token == remoteDev.Token {
		return true, nil
	}

	return false, nil
}

// getToken gets the login information for a developer and gets a new token.
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

		token, err = broome.GetTokenByLogin(email, pass)
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

// updateDeveloper gets the most up to date dev data and saves it.
func updateDeveloper(dev *db.Developer) error {
	// Get the developer from the devs token.
	developer, err := broome.GetDeveloper(dev.Token)
	if err != nil {
		return err
	}

	// Save the developer.
	log.Debug("Found developer", developer)
	dev.Developer = developer
	return dev.Save()
}
