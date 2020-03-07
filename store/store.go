package store

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

// Auth reads the auth specification from a local file.
func Auth() (*model.Auth, error) {
	authPath := os.Getenv(("AUTH_PATH"))
	if authPath == "" {
		authPath = "/app/auth.json"
	}
	bytes, err := ioutil.ReadFile(authPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read auth file '%s'", authPath)
	}
	var auth model.Auth
	err = json.Unmarshal(bytes, &auth)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal auth file '%s'", authPath)
	}

	return &auth, nil
}
