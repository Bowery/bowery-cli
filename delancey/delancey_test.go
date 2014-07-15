package delancey

// TODO (thebyrd) Delancey Methods rely on the .bowery/state file to be created. Not sure how to mock this out for the tests.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Bowery/bowery/broome"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/responses"
	"github.com/Bowery/gopackages/schemas"
)

var (
	TestDeveloper       *schemas.Developer
	TestApplications    []*schemas.Application
	TestState           *db.State
	TestImage           *schemas.Image
	globalTestDeveloper *schemas.Developer
)

var TestService = &schemas.Service{
	DockerID:      "92e91347f0ce8ab97823a8c6fb0b6d3e0424ef12e7c3ea2c6ebcf9206fd61cb6",
	Name:          "testservice",
	PrivateAddr:   "0.0.0.0:1",
	PublicAddr:    "0.0.0.0:2",
	SatelliteAddr: "0.0.0.0:3",
	SSHAddr:       "0.0.0.0:4",
	Image:         "testimage",
	CustomPorts:   map[string]string{"5": "127.0.0.1:2000"},
	Start:         "start cmd",
	Build:         "build cmd",
	Test:          "test cmd",
}

var TestVersion = "2.1.2"

func init() {
	err := os.Setenv("ENV", "testing")
	if err != nil {
		return
	}

	TestDeveloper, err = broome.CreateDeveloper("steve", "steve"+strconv.Itoa(time.Now().Nanosecond())+"@bowery.io", "somepassword")
	if err != nil {
		fmt.Println("Create test developer failed")
		fmt.Println(err)
		os.Exit(1)
	}

	TestApplications = []*schemas.Application{
		&schemas.Application{
			ID:          "5303a1636462d4d468000002",
			Name:        "someapp",
			DeveloperID: TestDeveloper.ID.Hex(),
			UpdatedAt:   1398102273378,
			Services:    []*schemas.Service{},
		},
	}

	TestState = &db.State{
		Token:  TestDeveloper.Token,
		App:    TestApplications[0],
		Config: map[string]*db.Service{},
		Path:   filepath.Join(".bowery", "state"),
	}

	TestState.Save()

	TestImage = &schemas.Image{
		ID:          "5303a1636462d4d468000003",
		Name:        "testimage",
		Description: "desc",
		CreatorID:   TestDeveloper.ID.Hex(),
		UpdatedAt:   1398102273378,
	}

	return
}
func TestUploadSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(uploadHandler))
	defer server.Close()

	addr, _ := url.Parse(server.URL)

	file, err := os.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = Upload(addr.Host, TestService.Name, file)
	if err != nil {
		t.Fatal(err)
	}

	os.Remove("test.txt")
}

func uploadHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "created"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestUpdateSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(updateHandler))
	defer server.Close()

	addr, _ := url.Parse(server.URL)

	file, err := os.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = Update(addr.Host, TestService.Name, "test.txt", "test.txt", "update")
	if err != nil {
		t.Fatal(err)
	}

	os.Remove("test.txt")
}

func updateHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "updated"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}
