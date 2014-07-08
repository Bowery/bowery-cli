// Copyright 2013-2014 Bowery, Inc.
// Package requests implements the api requests. It only exists to prevent
// a cyclic import for api and db.
package requests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/gopackages/schemas"
)

// GetTokenByLogin creates a token for the given devs email.
func GetTokenByLogin(email, password string) (string, error) {
	var body bytes.Buffer
	bodyReq := &LoginReq{Email: email, Password: password}

	encoder := json.NewEncoder(&body)
	err := encoder.Encode(bodyReq)
	if err != nil {
		return "", errors.NewStackError(err)
	}

	res, err := http.Post(api.BroomePath+api.CreateTokenPath, "application/json", &body)
	if err != nil {
		return "", errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	createRes := new(CreateTokenRes)
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

	res, err := http.Post(api.BroomePath+api.CreateDeveloperPath, "application/json", &body)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	createRes := new(DeveloperRes)
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
	res, err := http.Get(api.BroomePath + strings.Replace(api.MePath, "{token}", token, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	devRes := new(DeveloperRes)
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

// GetAppById retrieves a single application owned by the developer by id.
func GetAppById(id string) (*schemas.Application, error) {
	res, err := http.Get(api.BasePath + strings.Replace(api.AppPath, "{id}", id, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response
	appRes := new(AppRes)
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
	res, err := http.Get(api.BasePath + strings.Replace(api.AppsPath, "{token}", token, -1))
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response
	appsRes := new(AppsRes)
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

	endpoint := api.BasePath + strings.Replace(api.DestroyPath, "{id}", appID, -1)
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
	destroyRes := new(Res)
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
	res, err := http.Get(api.BasePath + api.VersionPath)
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

	log.Debug(api.BasePath + api.ConnectPath)
	res, err := http.Post(api.BasePath+api.ConnectPath, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	connectRes := new(ConnectRes)
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

	res, err := http.Post(api.BasePath+api.DisconnectPath, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	return nil
}

// RestartServce restarts the given service's docker container.
// RestartService restarts the given service's docker container.
func RestartService(dockerId, token string) (*schemas.Service, error) {
	endpoint := api.BasePath + strings.Replace(api.RestartPath, "{dockerid}", dockerId, -1)
	endpoint = strings.Replace(endpoint, "{token}", token, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	restartRes := new(RestartRes)
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
	endpoint := api.BasePath + strings.Replace(api.RemovePath, "{dockerid}", dockerId, -1)
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

	removeRes := new(Res)
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

	endpoint := api.BasePath + strings.Replace(api.SavePath, "{appid}", state.App.ID, -1)
	log.Debug(endpoint)
	res, err := http.Post(endpoint, "application/json", &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	saveRes := new(SaveRes)
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
	endpoint := api.BasePath + strings.Replace(api.BoweryImagesSearchPath, "{name}", name, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	imageRes := new(ImageTypeRes)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(imageRes)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	return imageRes.Images, nil
}

// FindImage checks if an image name exists, returning nil if found.
func FindImage(name string) error {
	endpoint := api.BasePath + strings.Replace(api.BoweryImagesCheckPath, "{name}", name, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	imageRes := new(Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(imageRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	if imageRes.Status == "found" {
		return nil
	}

	return errors.ErrNoImageFound
}

// Healthz checks to see if api is up and running.
func Healthz() error {
	res, err := http.Get(api.BasePath + api.HealthzPath)
	if err != nil {
		// Don't create stack error since we only use it to check connection.
		return err
	}
	defer res.Body.Close()

	return nil
}

// CheckHealth checks to see if the container is up
func SatelliteCheckHealth(url string) error {
	_, err := http.Get("http://" + url + "/healthz")

	return err
}

// DevPing checks to see if the api is up and running and updates
// the developers lastActive field.
func DevPing(token string) error {
	endpoint := api.BroomePath + strings.Replace(api.CheckPath, "{token}", token, -1)
	res, err := http.Get(endpoint)
	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	return nil
}

// SatelliteUpload uploads the given tar stream to the satellite endpoint.
func SatelliteUpload(url, serviceName string, file io.Reader) error {
	var body bytes.Buffer

	// Create the form body, and add the file field.
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "upload")
	if err != nil {
		return errors.NewStackError(err)
	}
	// Copy the file to the part.
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	// Get current app and add fields for init, build, test, and start.
	state, err := db.GetState()
	if err != nil {
		return err
	}

	if service := state.Config[serviceName]; service != nil {
		writer.WriteField("init", service.Init)
		writer.WriteField("build", service.Build)
		writer.WriteField("test", service.Test)
		writer.WriteField("start", service.Start)
	}

	if err = writer.Close(); err != nil {
		return err
	}

	res, err := http.Post("http://"+url, writer.FormDataContentType(), &body)
	if err != nil {
		if isRefusedConn(err) {
			err = errors.ErrSyncFailed
		}

		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	uploadRes := new(Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(uploadRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Created, so no error.
	if uploadRes.Status == "created" {
		return nil
	}

	return errors.NewStackError(uploadRes)
}

// SatelliteUpdate updates the given name with the status and path.
func SatelliteUpdate(url, serviceName, fullPath, name, status string) error {
	var body bytes.Buffer

	// Create writer, and write form fields.
	writer := multipart.NewWriter(&body)
	err := writer.WriteField("type", status)
	if err == nil {
		err = writer.WriteField("path", path.Join(strings.Split(name, string(filepath.Separator))...))
	}
	if err != nil {
		return errors.NewStackError(err)
	}

	// Attach file if update or create.
	if status == "update" || status == "create" {
		file, err := os.Open(fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		// Add file mode to write with.
		err = writer.WriteField("mode", strconv.FormatUint(uint64(stat.Mode().Perm()), 10))
		if err != nil {
			return errors.NewStackError(err)
		}

		part, err := writer.CreateFormFile("file", "upload")
		if err != nil {
			return errors.NewStackError(err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}

	// Get current app and add fields for init, build, test, and start.
	state, err := db.GetState()
	if err != nil {
		return err
	}

	if service := state.Config[serviceName]; service != nil {
		writer.WriteField("init", service.Init)
		writer.WriteField("build", service.Build)
		writer.WriteField("test", service.Test)
		writer.WriteField("start", service.Start)
	}

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", "http://"+url, &body)
	if err != nil {
		return errors.NewStackError(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if isRefusedConn(err) {
			err = errors.ErrSyncFailed
		}

		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	uploadRes := new(Res)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(uploadRes)
	if err != nil {
		return errors.NewStackError(err)
	}

	// Created, so no error.
	if uploadRes.Status == "updated" {
		return nil
	}

	return errors.NewStackError(uploadRes)
}

// SatelliteDownload retrieves the contents of a service.
func SatelliteDownload(url string) (io.Reader, error) {
	var buf bytes.Buffer
	res, err := http.Get("http://" + url)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		downloadRes := new(Res)
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(downloadRes)
		if err != nil {
			return nil, errors.NewStackError(err)
		}

		return nil, errors.NewStackError(downloadRes)
	}

	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	return &buf, nil
}

// DownloadNewVersion retrieves the version of bowery requested for the
// current os/arch.
func DownloadNewVersion(version string) (io.Reader, error) {
	var buf bytes.Buffer
	downloadPath := strings.Replace(api.DownloadPath, "{version}", version, -1)
	downloadPath = strings.Replace(downloadPath, "{os}", runtime.GOOS, -1)
	downloadPath = strings.Replace(downloadPath, "{arch}", runtime.GOARCH, -1)

	res, err := http.Get(downloadPath)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.NewStackError(errors.New(res.Status))
	}

	// Create a temp file to write zip to, the zip reader requires an
	// offset reader.
	temp, err := ioutil.TempFile("", "bowery_zip")
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer func() {
		temp.Close()
		os.RemoveAll(temp.Name())
	}()

	written, err := io.Copy(temp, res.Body)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	reader, err := zip.NewReader(temp, written)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	if len(reader.File) <= 0 {
		return nil, errors.NewStackError(errors.ErrVersionDownload)
	}

	// Get the first file.
	idx := -1
	for i, header := range reader.File {
		if header.FileInfo().IsDir() {
			continue
		}

		idx = i
		break
	}

	if idx < 0 {
		return nil, errors.NewStackError(errors.ErrVersionDownload)
	}

	// Assume the first file is the binary.
	file, err := reader.File[idx].Open()
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer file.Close()

	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, errors.NewStackError(err)
	}

	return &buf, nil
}

// isRefusedConn checks if a connection was refused.
func isRefusedConn(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// Another option is to inspect the error tree until we get to a syscall
	// but that is OS dependent so this is easier.
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "refused") {
		return true
	}

	return false
}

// ResetPassword sends request to broome to send password reset email
func ResetPassword(email string) error {
	res, err := http.Get(api.BroomePath + strings.Replace(api.ResetPasswordPath, "{email}", email, -1))

	if err != nil {
		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	resetRes := new(Res)
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
