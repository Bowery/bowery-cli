// Copyright 2013-2014 Bowery, Inc.
package db

import (
	"io"
	"os"
	"strconv"

	"github.com/Bowery/bowery/errors"
)

// Service describes a service.
type Service struct {
	Image string            `json:"image,omitempty"`
	Path  string            `json:"path,omitempty"`
	Ports []interface{}     `json:"ports,omitempty"`
	Start string            `json:"start,omitempty"`
	Build string            `json:"build,omitempty"`
	Test  string            `json:"test,omitempty"`
	Init  string            `json:"init,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
}

// Services contains a map of services by name.
type Services struct {
	Data map[string]*Service `json:"data"`
	Path string              `json:"-"`
}

// GetServices retrieves the services from the bowery.json file,
// and if empty uses empty services.
func GetServices() (*Services, error) {
	services := new(Services)
	services.Data = make(map[string]*Service)
	services.Path = "bowery.json"

	err := jumpToParent(services.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = load(&services.Data, services.Path)
	if err == io.EOF || os.IsNotExist(err) {
		err = nil
	}

	// Convert ports to ints.
	for _, service := range services.Data {
		if service.Ports != nil {
			for i, port := range service.Ports {
				_, isNum := port.(float64)
				portStr, isStr := port.(string)
				if !isNum && !isStr {
					return nil, errors.Newf(errors.ErrInvalidJSONTmpl, services.Path)
				}

				if isNum {
					continue
				}

				portNum, err := strconv.Atoi(portStr)
				if err != nil {
					return nil, errors.Newf(errors.ErrInvalidJSONTmpl, services.Path)
				}

				service.Ports[i] = portNum
			}
		}
	}

	return services, err
}

// Save writes the services to the bowery.json file.
func (services *Services) Save() error {
	return save(services.Data, services.Path)
}
