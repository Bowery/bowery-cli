// Copyright 2013-2014 Bowery, Inc.
// Package skylab implements the api requests. It only exists to prevent
// a cyclic import for api and db.
package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/responses"
	"github.com/Bowery/gopackages/log"
	"github.com/Bowery/gopackages/schemas"
)

// GetAppById retrieves a single application owned by the developer by id.
func GetAppById(id string) (*schemas.Application, error) {
	res, err := http.Get(BasePath + strings.Replace(AppPath, "{id}", id, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response
	appRes := new(responses.AppRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(appRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	// If found return application.
	if appRes.Status == "found" {
		return appRes.Application, nil
	}

	if strings.Contains(appRes.Error(), "No application found") {
		return nil, appRes
	}

	// Non "found" status indicates error.
	return nil, errors.NewStackError(appRes)
}

// GetApps retrieves the applications owned by the developer with the given token.
func GetApps(token string) ([]*schemas.Application, error) {
	res, err := http.Get(BasePath + strings.Replace(AppsPath, "{token}", token, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response
	appsRes := new(responses.AppsRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(appsRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	// If found return applications.
	if appsRes.Status == "found" {
		return appsRes.Applications, nil
	}

	if strings.Contains(appsRes.Error(), "Invalid Token") {
		return nil, errors.ErrInvalidToken
	}

	// Non "found" status indicates error.
	return nil, errors.NewStackError(appsRes)
}

// Destroy app sends a request to remove the app.
func DestroyAppByID(appID, token string) error {
	var body bytes.Buffer
	bodyReq := TokenReq{
		Token: token,
	}
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(bodyReq)
	if err != nil {
		return errors.NewStackError(err)
	}

	endpoint := BasePath + strings.Replace(DestroyPath, "{id}", appID, -1)
	req, err := http.NewRequest("PUT", endpoint, &body)
	if err != nil {
		return errors.NewStackError(err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Decode json response.
	destroyRes := new(responses.Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(destroyRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Success.
	if destroyRes.Status == "success" {
		return nil
	}

	return errors.NewStackError(destroyRes)
}

// GetVersion retrieves the latest version of the Bowery CLI.
func GetVersion() (string, error) {
	res, err := http.Get(BasePath + VersionPath)
	if err != nil {
		return "", errors.NewStackError(err)
	}
	defer res.Body.Close()

	version, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.NewStackError(err)
	}

	return string(version), nil
}

// Connect updates the given application state.
func Connect(state *db.State) error {
	// Encode state to json.
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(state)
	if err != nil {
		return errors.NewStackError(err)
	}

	log.Debug(BasePath + ConnectPath)
	res, err := http.Post(BasePath+ConnectPath, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	connectRes := new(responses.ConnectRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(connectRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Created, set the state.
	if connectRes.Status == "created" {
		state.App = connectRes.App
		state.Config = connectRes.Config
		return nil
	}

	if strings.Contains(connectRes.Error(), "Invalid Token") {
		return errors.ErrInvalidToken
	}

	if strings.Contains(connectRes.Error(), "hosts are unavailable") {
		return errors.NewStackError(errors.ErrOverCapacity)
	}

	// If user error, don't create stack.
	if strings.Contains(connectRes.Error(), "could not be found") ||
		strings.Contains(connectRes.Error(), "no longer valid") {
		return connectRes
	}

	// Non "created" status indicates error.
	return errors.NewStackError(connectRes)
}

func Disconnect(token string) error {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(map[string]string{"token": token})
	if err != nil {
		return errors.NewStackError(err)
	}

	res, err := http.Post(BasePath+DisconnectPath, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	return nil
}

// RestartService restarts the given service's docker container.
func RestartService(dockerId, token string) (*schemas.Service, error) {
	endpoint := BasePath + strings.Replace(RestartPath, "{dockerid}", dockerId, -1)
	endpoint = strings.Replace(endpoint, "{token}", token, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	restartRes := new(responses.RestartRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(restartRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	// Successfully restarted, return no error.
	if restartRes.Status == "success" {
		return restartRes.Service, nil
	}

	if strings.Contains(restartRes.Error(), "Invalid Token") {
		return nil, errors.ErrInvalidToken
	}

	// Non "created" status indicates error, just return invalid.
	return nil, errors.NewStackError(errors.New(errors.ErrFailedRestart, restartRes))
}

func RemoveService(dockerId, token string) error {
	endpoint := BasePath + strings.Replace(RemovePath, "{dockerid}", dockerId, -1)
	endpoint = strings.Replace(endpoint, "{token}", token, -1)

	var body bytes.Buffer
	reqBody := TokenReq{
		Token: token,
	}
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(reqBody)
	if err != nil {
		return errors.NewStackError(err)
	}

	req, err := http.NewRequest("DELETE", endpoint, &body)
	if err != nil {
		return errors.NewStackError(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	removeRes := new(responses.Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(removeRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Successfully restarted, return no error.
	if removeRes.Status == "success" {
		return nil
	}

	return errors.NewStackError(removeRes)
}

// SaveService saves an image of the specified service.
func SaveService(state *db.State, dev *db.Developer, serviceName, serviceAddr, imageName, imageDesc string) error {
	var body bytes.Buffer
	bodyReq := SaveServiceReq{
		Service:     serviceName,
		Image:       imageName,
		Description: imageDesc,
		App:         state.App,
		Config:      state.Config,
		Token:       dev.Token,
	}

	encoder := json.NewEncoder(&body)
	err := encoder.Encode(bodyReq)
	if err != nil {
		return errors.NewStackError(err)
	}

	endpoint := BasePath + strings.Replace(SavePath, "{appid}", state.App.ID, -1)
	log.Debug(endpoint)
	res, err := http.Post(endpoint, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	saveRes := new(responses.SaveRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(saveRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Created, set the state.
	if saveRes.Status == "created" {
		state.App = saveRes.App
		state.Config = saveRes.Config
		err = state.Save()
		if err != nil {
			return err
		}

		services, err := db.GetServices()
		if err != nil {
			return err
		}

		services.Data = saveRes.Config
		err = services.Save()
		if err != nil {
			return err
		}

		return nil
	} else if saveRes.Message == "Image already exists" {
		return errors.ErrImageExists
	}

	if strings.Contains(saveRes.Error(), "Invalid Token") {
		return errors.ErrInvalidToken
	}

	if strings.Contains(saveRes.Error(), "Image already exists") {
		return saveRes
	}

	// non 'created' status is handled as an error
	return errors.NewStackError(saveRes)
}

// SearchImages searches images by name
func SearchImages(name string) ([]*schemas.Image, error) {
	endpoint := BasePath + strings.Replace(BoweryImagesSearchPath, "{name}", name, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	imageRes := new(responses.ImageTypeRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(imageRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	return imageRes.Images, nil
}

// Healthz checks to see if api is up and running.
func Healthz() error {
	res, err := http.Get(BasePath + HealthzPath)
	if err != nil {
		// Don't create stack error since we only use it to check connection.
		return err
	}
	defer res.Body.Close()

	return nil
}

// DownloadNewVersion retrieves the version of bowery requested for the
// current os/arch.
// TODO (thebyrd) move this somewhere else
func DownloadNewVersion(version string) (io.Reader, io.Reader, error) {
	var binaryBuf bytes.Buffer
	var releaseNotesBuf bytes.Buffer

	downloadPath := strings.Replace(DownloadPath, "{version}", version, -1)
	downloadPath = strings.Replace(downloadPath, "{os}", runtime.GOOS, -1)
	downloadPath = strings.Replace(downloadPath, "{arch}", runtime.GOARCH, -1)

	res, err := http.Get(downloadPath)
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, nil, errors.NewStackError(errors.New(res.Status))
	}

	// Create a temp file to write zip to, the zip reader requires an
	// offset reader.
	temp, err := ioutil.TempFile("", "bowery_zip")
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}
	defer func() {
		temp.Close()
		os.RemoveAll(temp.Name())
	}()

	written, err := io.Copy(temp, res.Body)
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}

	reader, err := zip.NewReader(temp, written)
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}

	if len(reader.File) <= 0 {
		return nil, nil, errors.NewStackError(errors.ErrVersionDownload)
	}

	// Determine which file is the release notes and which
	// is the binary.

	binaryIndex := -1
	releaseNotesIndex := -1

	// Get the first file.
	for i, header := range reader.File {
		if header.FileInfo().IsDir() {
			continue
		}

		fileName := header.FileInfo().Name()
		if fileName == "bowery" || fileName == "bowery.exe" {
			binaryIndex = i
		} else if fileName == fmt.Sprintf("%s.txt", version) {
			releaseNotesIndex = i
		}

		if binaryIndex != -1 && releaseNotesIndex != -1 {
			break
		}
	}

	if binaryIndex < 0 || releaseNotesIndex < 0 {
		return nil, nil, errors.NewStackError(errors.ErrVersionDownload)
	}

	// Create buffer for binary.
	binary, err := reader.File[binaryIndex].Open()
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}
	defer binary.Close()

	_, err = io.Copy(&binaryBuf, binary)
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}

	// Create buffer for release notes.
	release, err := reader.File[releaseNotesIndex].Open()
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}
	defer release.Close()

	_, err = io.Copy(&releaseNotesBuf, release)
	if err != nil {
		return nil, nil, errors.NewStackError(err)
	}

	return &binaryBuf, &releaseNotesBuf, nil
}
