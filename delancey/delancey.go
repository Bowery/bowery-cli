package delancey

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/responses"
)

// Download retrieves the contents of a service.
func Download(url string) (io.Reader, error) {
	var buf bytes.Buffer
	res, err := http.Get("http://" + url)
	if err != nil {
		return nil, errors.NewStackError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		downloadRes := new(responses.Res)
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

// Upload sends an upload request to a satellite endpoint, including
// a tar upload file if given.
func Upload(url, serviceName string, file *os.File) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Write file to multipart body if given.
	if file != nil {
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

	service := state.Config[serviceName]
	if service != nil {
		err = writer.WriteField("init", service.Init)
		if err == nil {
			err = writer.WriteField("build", service.Build)
		}
		if err == nil {
			err = writer.WriteField("test", service.Test)
		}
		if err == nil {
			err = writer.WriteField("start", service.Start)
		}
	}
	if err == nil {
		err = writer.Close()
	}
	if err != nil {
		return err
	}

	res, err := http.Post("http://"+url, writer.FormDataContentType(), &body)
	if err != nil {
		if responses.IsRefusedConn(err) {
			err = errors.ErrSyncFailed
		}

		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	uploadRes := new(responses.Res)
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

// Update updates the given name with the status and path.
func Update(url, serviceName, fullPath, name, status string) error {
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
		if responses.IsRefusedConn(err) {
			err = errors.ErrSyncFailed
		}

		return errors.NewStackError(err)
	}
	defer res.Body.Close()

	// Decode json response.
	uploadRes := new(responses.Res)
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

// CheckHealth checks to see if the container is up
func CheckHealth(url string) error {
	_, err := http.Get("http://" + url + "/healthz")
	return err
}
