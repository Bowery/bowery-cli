// Copyright 2013-2014 Bowery, Inc.
package rollbar

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var client = &Client{
	Token: "testtoken",
	Env:   "testing",
}

func init() {
	return
}

func TestReportSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(reportHandler))
	defer server.Close()
	Endpoint = server.URL + "/"

	err := client.Report(errors.New("error"))
	if err != nil {
		t.Fatal(err)
	}
}

func reportHandler(rw http.ResponseWriter, req *http.Request) {
	res := &Response{Message: "success", Code: 0}
	body, _ := json.Marshal(res)
	rw.Write(body)
}
