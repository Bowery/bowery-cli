// Copyright 2013-2014 Bowery, Inc.
package api

import (
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/gopackages/schemas"
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
