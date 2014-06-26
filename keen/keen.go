// Copyright 2013-2014 Bowery, Inc.
// Package keen implements routines to add events to the keen API.
package keen

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// The base keen url including version.
var BaseURL = "https://api.keen.io/3.0/projects/"

// Error for when the response is empty.
var ErrEmptyRes = errors.New("keen: Empty response from server.")

// Response contains the fields keen responds with.
type Response struct {
	Message string `json:"message"`
	ErrCode string `json:"error_code"`
}

func (res *Response) Error() string {
	return "keen(" + res.ErrCode + "): " + res.Message
}

// Client contains info to authenticate with the keen API.
type Client struct {
	WriteKey  string
	ProjectID string
}

// Add adds a new collection to the event given.
func (client *Client) AddEvent(event string, collection interface{}) error {
	var body bytes.Buffer

	encoder := json.NewEncoder(&body)
	err := encoder.Encode(collection)
	if err != nil {
		return err
	}

	// Create request adding authorization token and setting content info.
	req, err := http.NewRequest("POST", BaseURL+client.ProjectID+"/events/"+event, &body)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", client.WriteKey)
	req.Header.Add("Content-Type", "application/json")
	req.ContentLength = int64(body.Len())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Response was successful.
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	// Decode response to retrieve error info.
	resBody := new(Response)
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(resBody)
	if err != nil {
		if err == io.EOF {
			return ErrEmptyRes
		}

		return err
	}

	return resBody
}
