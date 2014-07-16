// Copyright 2013-2014 Bowery, Inc.
// Package api contains hosts for bowery, redis, and contains the various
// endpoints.
package api

import (
	"os"

	"github.com/Bowery/bowery/db"
)

var (
	env         = os.Getenv("ENV")
	host        = os.Getenv("HOST")
	boweryApi   = os.Getenv("API_ADDR")
	boweryRedis = os.Getenv("REDIS_ADDR")
)

// Base endpoints for api and Redis.
var (
	BasePath  = "http://api.bowery.io"
	RedisPath = "ec2-23-22-237-84.compute-1.amazonaws.com:6379"
)

// Paths that are used to call api endpoints.
const (
	AppPath                = "/applications/{id}"
	DestroyPath            = "/applications/{id}/destroy"
	AppsPath               = "/developers/applications?token={token}"
	ConnectPath            = "/connect"
	DisconnectPath         = "/disconnect"
	VersionPath            = "/version/cli"
	HealthzPath            = "/healthz"
	RestartPath            = "/services/{dockerid}/restart?token={token}"
	SavePath               = "/services/{appid}/save"
	RemovePath             = "/services/{dockerid}/remove?token={token}"
	DownloadPath           = "http://download.bowery.io/{version}_{os}_{arch}.zip"
	BoweryImagesSearchPath = "/images/search/{name}"
	BoweryImagesCheckPath  = "/images/{name}"
)

func init() {
	if env == "development" {
		if host == "" {
			host = "localhost"
		}

		BasePath = "http://" + host + ":3000"
		RedisPath = host + ":6379"

		if boweryApi != "" {
			BasePath = "http://" + boweryApi
		}

		if boweryRedis != "" {
			RedisPath = boweryRedis
		}
	}

	dev, _ := db.GetDeveloper()
	if dev != nil && dev.Config != nil {
		h, ok := dev.Config["host"]
		if ok && h != "" {
			BasePath = h
		}

		redis, ok := dev.Config["redis"]
		if ok && redis != "" {
			RedisPath = redis
		}
	}
}
