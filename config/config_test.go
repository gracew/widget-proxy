package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	createdByAuthPolicy := model.AuthPolicy{
		Type: model.AuthPolicyTypeCreatedBy,
	}
	input := model.Auth{
		APIID: "apiID",
		Read:  &createdByAuthPolicy,
	}

	path, err := writeTmpFile(input, "auth-")
	assert.NoError(t, err)

	output, err := Auth(path)
	assert.NoError(t, err)
	assert.Equal(t, input, *output)
}

func TestCustomLogic(t *testing.T) {
	beforeSave := "before"
	afterSave := "after"
	customLogic1 := &model.CustomLogic{
		Before: &beforeSave,
	}
	customLogic2 := &model.CustomLogic{
		After: &afterSave,
	}
	input := model.AllCustomLogic{APIID: "apiID", Create: customLogic1, Delete: customLogic2}

	path, err := writeTmpFile(input, "custom-logic-")
	assert.NoError(t, err)

	output, err := CustomLogic(path)
	assert.NoError(t, err)
	assert.Equal(t, input, *output)
}

func writeTmpFile(input interface{}, prefix string) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}

	err = json.NewEncoder(file).Encode(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode object to file")
	}
	return filepath.Abs(file.Name())
}
