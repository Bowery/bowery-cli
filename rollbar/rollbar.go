// Copyright 2013-2014 Bowery, Inc.
// Package rollbar contains routines to report errors to the rollbar API.
package rollbar

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/version"
)

var debug = os.Getenv("DEBUG")

// The endpoint errors are reported to.
var Endpoint = "https://api.rollbar.com/api/1/item/"

// Response is a typical response.
type Response struct {
	Code    int    `json:"err"`
	Message string `json:"message"`
}

func (res *Response) Error() string {
	return "rollbar(" + strconv.Itoa(res.Code) + "): " + res.Message
}

// Request defines the request body or the rollbar API.
type Request struct {
	AccessToken string `json:"access_token"`
	Data        *Data  `json:"data"`
}

// Data contains fields that define the error and it's environment.
type Data struct {
	Environment string  `json:"environment"`
	Body        *body   `json:"body"`
	Level       string  `json:"level"`
	Timestamp   int64   `json:"timestamp"`
	Platform    string  `json:"platform"`
	Language    string  `json:"language"`
	Title       string  `json:"title"`
	Server      *server `json:"server"`
	Custom      *custom `json:"custom"`
}

// body contains a stack trace for the request.
type body struct {
	Trace *errors.Trace `json:"trace"`
}

// server contains the systems host.
type server struct {
	Host string `json:"host"`
}

// custom contains various extra information for the request.
type custom struct {
	Arch     string       `json:"arch"`
	ID       string       `json:"id,omitempty"`
	Email    string       `json:"email,omitempty"`
	Name     string       `json:"name,omitempty"`
	State    *db.State    `json:"state,omitempty"`
	Services *db.Services `json:"services,omitempty"`
	Version  string       `json:"version"`
}

// Client contains info for reporting to the given client token.
type Client struct {
	Token string
	Env   string
}

// Report sends an error message, and prints information to the user.
func (client *Client) Report(err error) error {
	if err == nil {
		return nil
	}

	if err != errors.ErrCTRLC {
		log.Fprintln(os.Stderr, "red", errors.ErrContactSupport)
	}
	log.Fprintln(os.Stderr, "red", err)

	// Don't report non stack errors.
	se, ok := err.(*errors.StackError)
	if !ok {
		return nil
	}

	// If development print the stack info.
	if client.Env == "development" {
		log.Fprintln(os.Stderr, "", "\n"+se.Stack())
		return nil
	}

	// If in debug mode, do not send to rollbar.
	if debug == "cli" {
		return nil
	}

	// Gather various details about the environment.
	services, _ := db.GetServices()
	state, _ := db.GetState()
	host, _ := os.Hostname()

	reqBody := &Request{
		AccessToken: client.Token,
		Data: &Data{
			Environment: client.Env,
			Body:        &body{Trace: se.Trace},
			Level:       "error",
			Timestamp:   time.Now().Unix(),
			Platform:    runtime.GOOS,
			Language:    "go",
			Title:       "Error Report: " + err.Error(),
			Server:      &server{Host: host},
			Custom: &custom{
				Arch:     runtime.GOARCH,
				Services: services,
				State:    state,
				Version:  version.Version,
			},
		},
	}

	dev, _ := db.GetDeveloper()
	if dev != nil && dev.Developer != nil {
		reqBody.Data.Custom.ID = dev.Developer.ID.Hex()
		reqBody.Data.Custom.Email = dev.Developer.Email
		reqBody.Data.Custom.Name = dev.Developer.Name
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	err = encoder.Encode(reqBody)
	if err != nil {
		return err
	}

	res, err := http.Post(Endpoint, "application/json", &body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Response was successful.
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	// Decode respsonse to retrieve error info.
	resBody := new(Response)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(resBody)
	if err != nil {
		return err
	}

	return resBody
}
