// Copyright 2013-2014 Bowery, Inc.
package responses

import (
	"strings"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/gopackages/schemas"
)

// Res is the generic response with status and error message.
type Res struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

func (res *Res) Error() string {
	return res.Err
}

// CreateTokenRes contains the new token.
type CreateTokenRes struct {
	*Res
	Token string `json:"token"`
}

// DeveloperRes contains the returned developer.
type DeveloperRes struct {
	*Res
	Developer *schemas.Developer `json:"developer"`
}

// AppRes contains the returned apps.
type AppRes struct {
	*Res
	Application *schemas.Application `json:"application"`
}

// AppsRes contains the returned apps.
type AppsRes struct {
	*Res
	Applications []*schemas.Application `json:"applications"`
}

// ConnectRes contains the state returned from a connection.
type ConnectRes struct {
	*Res
	*db.State
}

// SaveRes contains the state and a save message.
type SaveRes struct {
	*Res
	*db.State
	Message string `json:"message"`
}

// ImageTypeRes contains a slice of images.
type ImageTypeRes struct {
	*Res
	Images []*schemas.Image `json:"images"`
}

// RestartRes contains a new service.
type RestartRes struct {
	*Res
	Service *schemas.Service `json:"service"`
}

// isRefusedConn checks if a connection was refused.
func IsRefusedConn(err error) bool {
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
