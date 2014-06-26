// Copyright 2013-2014 Bowery, Inc.
// Package schemas contains the structures needed to store and read data,
// from the api, satellite, and other services..
package schemas

// Developer is a developer from api.
type Developer struct {
	ID        string `json:"_id"`
	CreatedAt int64  `json:"createdAt"`
	Email     string `json:"email"`
	IsPaid    bool   `json:"isPaid"`
	License   string `json:"license"`
	Name      string `json:"name"`
	Token     string `json:"token"`
}

// Service is a service from api.
type Service struct {
	DockerID      string            `json:"dockerId"`
	Name          string            `json:"name"`
	PrivateAddr   string            `json:"privateAddr"`
	PublicAddr    string            `json:"publicAddr"`
	SatelliteAddr string            `json:"satelliteAddr"`
	SSHAddr       string            `json:"sshAddr"`
	Image         string            `json:"image"`
	CustomPorts   map[string]string `json:"customPorts"`
	Start         string            `json:"start,omitempty"`
	Build         string            `json:"build,omitempty"`
	Test          string            `json:"test,omitempty"`
	Init          string            `json:"init,omitempty"`
	Type          string            `json:"type,omitempty"` // incase there's bad data
}

// Application is an application from api.
type Application struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name,omitempty"`
	DeveloperID string     `json:"developerId"`
	UpdatedAt   int64      `json:"updatedAt"`
	Services    []*Service `json:"services"`
	IsActive    bool       `json:"isActive"`
}

// Image is an image from the api.
type Image struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   string `json:"creatorId"`
	UpdatedAt   int64  `json:"updatedAt"`
}
