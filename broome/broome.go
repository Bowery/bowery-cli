// Copyright 2013-2014 Bowery, Inc.
package broome

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/responses"
	"github.com/Bowery/gopackages/schemas"
)

var (
	Host    = "http://broome.io"
	hostEnv = os.Getenv("BROOME_ADDR")
	env     = os.Getenv("ENV")
)

const (
	CreateTokenPath     = "/developers/token"
	CreateDeveloperPath = "/developers"
	MePath              = "/developers/me?token={token}"
	CheckPath           = "/developers/me/check?token={token}"
	ResetPasswordPath   = "/reset/{email}"
)

func init() {
	if env == "development" {
		Host = "http://127.0.0.1:4000"
	}

	if hostEnv != "" {
		Host = "http://" + hostEnv
	}
}

// GetTokenByLogin creates a token for the given devs email.
func GetTokenByLogin(email, password string) (string, error) {
	var body bytes.Buffer
	bodyReq := &LoginReq{Email: email, Password: password}

	encoder := json.NewEncoder(&body)
	err := encoder.Encode(bodyReq)
	if err != nil {
		return "", errors.NewStackError(err)
	}

	res, err := http.Post(Host+CreateTokenPath, "application/json", &body)
	if err != nil {
		return "", errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	createRes := new(responses.CreateTokenRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(createRes)
	if err != nil {
		return "", errors.NewStackError(err)
	}

	// Created, just return token.
	if createRes.Status == "created" {
		return createRes.Token, nil
	}

	// Check for license issues.
	if strings.Contains(createRes.Error(), "License expired.") ||
		strings.Contains(createRes.Error(), "License user limit reached.") {
		return "", createRes
	}

	// Non "created" status indicates error, just return invalid.
	return "", errors.ErrInvalidLogin
}

// CreateDeveloper creates a new developer.
func CreateDeveloper(name, email, password string) (*schemas.Developer, error) {
	var body bytes.Buffer
	bodyReq := &LoginReq{Name: name, Email: email, Password: password}

	encoder := json.NewEncoder(&body)
	err := encoder.Encode(bodyReq)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	res, err := http.Post(Host+CreateDeveloperPath, "application/json", &body)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	createRes := new(responses.DeveloperRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(createRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	// Created, just return token.
	if createRes.Status == "created" {
		return createRes.Developer, nil
	}

	// If the error is about developer existing, don't create stack error.
	if strings.Contains(createRes.Error(), "email already exists") {
		return nil, errors.ErrDeveloperExists
	}

	// Check for license issues.
	if strings.Contains(createRes.Error(), "License expired.") ||
		strings.Contains(createRes.Error(), "License user limit reached.") {
		return nil, createRes
	}

	// Non "created" status indicates error, just return invalid.
	return nil, errors.NewStackError(createRes)
}

// GetDeveloper retrieves the developer for the given token.
func GetDeveloper(token string) (*schemas.Developer, error) {
	res, err := http.Get(Host + strings.Replace(MePath, "{token}", token, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	devRes := new(responses.DeveloperRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(devRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	// Found so return the developer.
	if devRes.Status == "found" {
		return devRes.Developer, nil
	}

	if strings.Contains(devRes.Error(), "Invalid Token") {
		return nil, errors.ErrInvalidToken
	}

	// Non "found" status indicates error.
	return nil, errors.NewStackError(devRes)
}

// ResetPassword sends request to broome to send password reset email
func ResetPassword(email string) error {
	res, err := http.Get(Host + strings.Replace(ResetPasswordPath, "{email}", email, -1))

	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	resetRes := new(responses.Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(resetRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	if resetRes.Status == "success" {
		return nil
	}

	if strings.Contains(resetRes.Error(), "not found") {
		return errors.ErrInvalidEmail
	}

	return errors.NewStackError(errors.ErrResetRequest)
}

// DevPing checks to see if the api is up and running and updates
// the developers lastActive field.
func DevPing(token string) error {
	endpoint := Host + strings.Replace(CheckPath, "{token}", token, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	return nil
}
