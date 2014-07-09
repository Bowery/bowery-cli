package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
			DeveloperID: TestDeveloper.ID,
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
		CreatorID:   TestDeveloper.ID,
		UpdatedAt:   1398102273378,
	}

	return
}

func GetAppByIdSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getAppByIdHandler))
	defer server.Close()
	BasePath = server.URL

	app, err := GetAppById("5303a1636462d4d468000002")
	if err != nil {
		t.Fatal(err)
	}

	if app.ID != "5303a1636462d4d468000002" {
		t.Error("Failed to get application")
	}
}

func getAppByIdHandler(rw http.ResponseWriter, req *http.Request) {
	res := &responses.AppRes{Res: &responses.Res{Status: "found"}, Application: TestApplications[0]}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func GetAppsSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getAppsHandler))
	defer server.Close()
	BasePath = server.URL

	dev := TestDeveloper

	apps, err := GetApps(dev.Token)
	if err != nil {
		t.Fatal(err)
	}

	for _, app := range apps {
		if app.DeveloperID != dev.ID {
			t.Error("Application does not belong to developer.")
		}
	}
}

func getAppsHandler(rw http.ResponseWriter, req *http.Request) {
	res := &responses.AppsRes{Res: &responses.Res{Status: "found"}, Applications: TestApplications}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestGetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getVersionHandler))
	defer server.Close()
	BasePath = server.URL

	version, err := GetVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != TestVersion {
		t.Error("Wrong version returned")
	}
}

func getVersionHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(TestVersion))
}

func TestDisconnectSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(disconnectHandler))
	defer server.Close()
	BasePath = server.URL

	dev := TestDeveloper

	if err := Disconnect(dev.Token); err != nil {
		t.Fatal(err)
	}
}

func disconnectHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestRestartServiceSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(restartServiceHandler))
	defer server.Close()
	BasePath = server.URL

	dev := TestDeveloper
	service := TestService

	if _, err := RestartService(service.DockerID, dev.Token); err != nil {
		t.Fatal(err)
	}
}

func restartServiceHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestRemoveServiceSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(removeServiceHandler))
	defer server.Close()
	BasePath = server.URL

	dev := TestDeveloper
	service := TestService

	if err := RemoveService(service.DockerID, dev.Token); err != nil {
		t.Fatal(err)
	}
}

func removeServiceHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

// TODO(steve): TestSaveServiceSuccessful

func TestSearchImagesSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(searchImagesHandler))
	defer server.Close()
	BasePath = server.URL

	images, err := SearchImages("testimage")
	if err != nil {
		t.Fatal(err)
	}

	if len(images) == 0 {
		t.Error("Response should not be successful if no images exist.")
	}
}

func searchImagesHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.ImageTypeRes{Res: &responses.Res{Status: "found"}, Images: []*schemas.Image{TestImage}}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestFindImageSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(findImageHandler))
	defer server.Close()
	BasePath = server.URL

	err := FindImage("testimage")
	if err != nil {
		t.Fatal(err)
	}
}

func findImageHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "found"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestHealthzSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(healthzHandler))
	defer server.Close()
	BasePath = server.URL

	err := Healthz()
	if err != nil {
		t.Fatal(err)
	}
}

func healthzHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("ok"))
}
