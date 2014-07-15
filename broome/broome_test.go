package broome

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

var TestVersion = "2.1.2"

func init() {
	err := os.Setenv("ENV", "testing")
	if err != nil {
		return
	}

	TestDeveloper, err = CreateDeveloper("steve", "steve"+strconv.Itoa(time.Now().Nanosecond())+"@bowery.io", "somepassword")
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

func TestGetTokenByLoginSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getTokenByLoginHandler))
	defer server.Close()
	Host = server.URL

	token, err := GetTokenByLogin(TestDeveloper.Email, "somepassword")
	TestDeveloper.Token = token
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("Did not get token.")
	}
}

func getTokenByLoginHandler(rw http.ResponseWriter, req *http.Request) {
	res := &responses.CreateTokenRes{Res: &responses.Res{Status: "created"}, Token: "sometoken"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestCreateDeveloperSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(createDeveloperHandler))
	defer server.Close()
	Host = server.URL

	email := "ricky" + strconv.Itoa(time.Now().Nanosecond()) + "@bowery.io"
	dev, err := CreateDeveloper("steve", email, "supersecurepassword")
	if err != nil {
		t.Fatal(err)
	}

	if dev.Name != TestDeveloper.Name || dev.Email != TestDeveloper.Email {
		t.Error("Failed to create new developer")
	}
}

func createDeveloperHandler(rw http.ResponseWriter, req *http.Request) {
	res := &responses.DeveloperRes{Res: &responses.Res{Status: "created"}, Developer: TestDeveloper}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestGetDeveloperSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(getDeveloperByTokenHandler))
	defer server.Close()
	Host = server.URL

	dev, err := GetDeveloper(TestDeveloper.Token)
	if err != nil {
		t.Fatal(err)
	}

	if dev.Token != TestDeveloper.Token {
		t.Error("Failed to get developer")
	}
}

func getDeveloperByTokenHandler(rw http.ResponseWriter, req *http.Request) {
	res := &responses.DeveloperRes{Res: &responses.Res{Status: "found"}, Developer: TestDeveloper}
	body, _ := json.Marshal(res)
	rw.Write(body)
}

func TestDevPingSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(devPingHandler))
	defer server.Close()
	Host = server.URL

	dev := TestDeveloper

	err := DevPing(dev.Token)
	if err != nil {
		t.Fatal(err)
	}
}

func devPingHandler(rw http.ResponseWriter, req *http.Request) {
	res := responses.Res{Status: "updated"}
	body, _ := json.Marshal(res)
	rw.Write(body)
}
