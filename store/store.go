package store

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

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
func CustomLogic(path string) ([]model.CustomLogic, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read custom logic file '%s'", path)
	}
	var customLogic []model.CustomLogic
	err = json.Unmarshal(bytes, &customLogic)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal custom logic file '%s'", path)
	}

	return customLogic, nil
}
