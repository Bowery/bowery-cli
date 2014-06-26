// Copyright 2013-2014 Bowery, Inc.
package keen

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var client = &Client{
	WriteKey:  "testkey",
	ProjectID: "testid",
}

func init() {
	return
}

func TestAddEventSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(addEventHandler))
	defer server.Close()
	BaseURL = server.URL + "/"

	err := client.AddEvent("testevent", "testdata")
	if err != nil {
		t.Fatal(err)
	}
}

func addEventHandler(rw http.ResponseWriter, req *http.Request) {
	res := &Response{Message: "success", ErrCode: "0"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}
