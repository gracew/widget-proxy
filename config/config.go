package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

const (
	ParseURL        = "http://parse:1337/parse/"
	PostgresAddress = "api-postgres:5432"
	CustomLogicURL  = "http://custom-logic:8080/"

	APIPath         = "/app/api.json"
	AuthPath        = "/app/auth.json"
	CustomLogicPath = "/app/customLogic.json"
)

var (
	APIName = os.Getenv("API_NAME")
)

// API reads the API specification from the given file.
func API(path string) (*model.API, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read api file '%s'", path)
	}
	var api model.API
	err = json.Unmarshal(bytes, &api)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal api file '%s'", path)
	}

	return &api, nil
}

// Auth reads the auth specification from the given file.
func Auth(path string) (*model.Auth, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read auth file '%s'", path)
	}
	var auth model.Auth
	err = json.Unmarshal(bytes, &auth)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal auth file '%s'", path)
	}

	return &auth, nil
}

// CustomLogic reads the custom logic specification from the given file.
func CustomLogic(path string) (*model.AllCustomLogic, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read custom logic file '%s'", path)
	}
	var customLogic model.AllCustomLogic
	err = json.Unmarshal(bytes, &customLogic)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal custom logic file '%s'", path)
	}

	return &customLogic, nil
}
