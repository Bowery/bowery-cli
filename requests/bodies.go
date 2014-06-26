// Copyright 2013-2014 Bowery, Inc.
package requests

import (
	"github.com/Bowery/cli/db"
	"github.com/Bowery/cli/schemas"
)

// TokenReq is a generic request containing only a developer token.
type TokenReq struct {
	Token string `json:"token"`
}

// SaveServiceReq contains the request body for the SaveService endpoint.
type SaveServiceReq struct {
	Service     string                 `json:"service"`
	Image       string                 `json:"image"`
	Description string                 `json:"description"`
	App         *schemas.Application   `json:"app"`
	Config      map[string]*db.Service `json:"config"`
	Token       string                 `json:"token"`
}

// BuildReq contains the request body for the BuildApp endpoint.
type BuildAppReq struct {
	Provider string               `json:"provider"`
	App      *schemas.Application `json:"app"`
	Token    string               `json:"token"`
}

// LoginReq contains fields for request bodies that user common login info
// such as email/password.
type LoginReq struct {
	Name     string `json:"name,omitempty"` // Only some use name.
	Email    string `json:"email"`
	Password string `json:"password"`
}
