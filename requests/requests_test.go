package requests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/Bowery/cli/api"
	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/schemas"
)

var TestDeveloper = &schemas.Developer{
	ID:        "5303a1636462d4d468000001",
	CreatedAt: 1398102273377,
	Email:     "steve@bowery.io",
	IsPaid:    true,
	License:   "somelicense",
	Name:      "steve",
	Token:     "sometoken",
}

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

var TestApplications = []*schemas.Application{
	&schemas.Application{
		ID:          "5303a1636462d4d468000002",
		Name:        "someapp",
		DeveloperID: TestDeveloper.ID,
		UpdatedAt:   1398102273378,
		Services:    []*schemas.Service{},
	},
}

var TestState = &db.State{
	Token:  TestDeveloper.Token,
	App:    TestApplications[0],
	Config: map[string]*db.Service{},
	Path:   filepath.Join(".bowery", "state"),
}

var TestImage = &schemas.Image{
	ID:          "5303a1636462d4d468000003",
	Name:        "testimage",
	Description: "desc",
	CreatorID:   TestDeveloper.ID,
	UpdatedAt:   1398102273378,
}

var TestVersion = "2.1.2"

func init() {
	err := os.Setenv("ENV", "testing")
	if err != nil {
		return
	}

	TestState.Save()

	return
}

func TestGetTokenByLoginSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getTokenByLoginHandler))
	defer server.Close()
	api.BasePath = server.URL

	token, err := GetTokenByLogin("cash@bowery.io", "supersecurepassword")
	if err != nil {
		t.Fatal(err)
	}

	if token != "sometoken" {
		t.Error("Token isn't the expected value", token)
	}
}

func getTokenByLoginHandler(rw http.ResponseWriter, req *http.Request) {
	res := &CreateTokenRes{Res: &Res{Status: "created"}, Token: "sometoken"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestCreateDeveloperSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(createDeveloperHandler))
	defer server.Close()
	api.BasePath = server.URL

	dev, err := CreateDeveloper("steve", "steve@bowery.io", "supersecurepassword")
	if err != nil {
		t.Fatal(err)
	}

	if dev.Name != "steve" || dev.Email != "steve@bowery.io" {
		t.Error("Failed to create new developer")
	}
}

func createDeveloperHandler(rw http.ResponseWriter, req *http.Request) {
	res := &DeveloperRes{Res: &Res{Status: "created"}, Developer: TestDeveloper}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestGetDeveloperSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getDeveloperByTokenHandler))
	defer server.Close()
	api.BasePath = server.URL

	dev, err := GetDeveloper("sometoken")
	if err != nil {
		t.Fatal(err)
	}

	if dev.Token != "sometoken" {
		t.Error("Failed to get developer")
	}
}

func getDeveloperByTokenHandler(rw http.ResponseWriter, req *http.Request) {
	res := &DeveloperRes{Res: &Res{Status: "found"}, Developer: TestDeveloper}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func GetAppByIdSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getAppByIdHandler))
	defer server.Close()
	api.BasePath = server.URL

	app, err := GetAppById("5303a1636462d4d468000002")
	if err != nil {
		t.Fatal(err)
	}

	if app.ID != "5303a1636462d4d468000002" {
		t.Error("Failed to get application")
	}
}

func getAppByIdHandler(rw http.ResponseWriter, req *http.Request) {
	res := &AppRes{Res: &Res{Status: "found"}, Application: TestApplications[0]}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func GetAppsSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getAppsHandler))
	defer server.Close()
	api.BasePath = server.URL

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
	res := &AppsRes{Res: &Res{Status: "found"}, Applications: TestApplications}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestGetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getVersionHandler))
	defer server.Close()
	api.BasePath = server.URL

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
	api.BasePath = server.URL

	dev := TestDeveloper

	if err := Disconnect(dev.Token); err != nil {
		t.Fatal(err)
	}
}

func disconnectHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestRestartServiceSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(restartServiceHandler))
	defer server.Close()
	api.BasePath = server.URL

	dev := TestDeveloper
	service := TestService

	if _, err := RestartService(service.DockerID, dev.Token); err != nil {
		t.Fatal(err)
	}
}

func restartServiceHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestRemoveServiceSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(removeServiceHandler))
	defer server.Close()
	api.BasePath = server.URL

	dev := TestDeveloper
	service := TestService

	if err := RemoveService(service.DockerID, dev.Token); err != nil {
		t.Fatal(err)
	}
}

func removeServiceHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "success"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

// TODO(steve): TestSaveServiceSuccessful

func TestSearchImagesSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(searchImagesHandler))
	defer server.Close()
	api.BasePath = server.URL

	images, err := SearchImages("testimage")
	if err != nil {
		t.Fatal(err)
	}

	if len(images) == 0 {
		t.Error("Response should not be successful if no images exist.")
	}
}

func searchImagesHandler(rw http.ResponseWriter, req *http.Request) {
	res := ImageTypeRes{Res: &Res{Status: "found"}, Images: []*schemas.Image{TestImage}}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestFindImageSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(findImageHandler))
	defer server.Close()
	api.BasePath = server.URL

	err := FindImage("testimage")
	if err != nil {
		t.Fatal(err)
	}
}

func findImageHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "found"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestHealthzSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(healthzHandler))
	defer server.Close()
	api.BasePath = server.URL

	err := Healthz()
	if err != nil {
		t.Fatal(err)
	}
}

func healthzHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("ok"))
}

func TestDevPingSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(devPingHandler))
	defer server.Close()
	api.BasePath = server.URL

	dev := TestDeveloper

	err := DevPing(dev.Token)
	if err != nil {
		t.Fatal(err)
	}
}

func devPingHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "updated"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestSatelliteUploadSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(satelliteUploadHandler))
	defer server.Close()

	addr, _ := url.Parse(server.URL)

	file, err := os.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = SatelliteUpload(addr.Host, TestService.Name, file)
	if err != nil {
		t.Fatal(err)
	}

	os.Remove("test.txt")
}

func satelliteUploadHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "created"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestSatelliteUpdateSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(satelliteUpdateHandler))
	defer server.Close()

	addr, _ := url.Parse(server.URL)

	file, err := os.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = SatelliteUpdate(addr.Host, TestService.Name, "test.txt", "test.txt", "update")
	if err != nil {
		t.Fatal(err)
	}

	os.Remove("test.txt")
}

func satelliteUpdateHandler(rw http.ResponseWriter, req *http.Request) {
	res := Res{Status: "updated"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}
