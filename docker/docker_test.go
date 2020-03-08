package docker

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// not really a test, but it spins up a node-runner instance that can be manually interacted with
func TestNode(t *testing.T) {
	buildCmd := exec.Command("docker",
		"build",
		"node",
		"-t",
		"node-runner",
	)
	err := buildCmd.Run()
	assert.NoError(t, err)

	beforeSave := `
function beforeSave(input) {
	input.concat = input.name + " " + input.score;
	return input;
}

module.exports = beforeSave;
`
	customLogic := model.CustomLogic{
		APIID:              "apiID",
		OperationType: model.OperationTypeCreate,
		BeforeSave: &beforeSave,
	}

	customLogicPath, err := writeTmpFile([]model.CustomLogic{customLogic}, "custom-logic-")
	assert.NoError(t, err)

	containerName := uuid.New().String()
	runCmd := exec.Command("docker",
		"run",
		"-d",
		"-v",
		customLogicPath + ":/app/customLogic.json",
		"-p",
		"9090:8080",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		"node-runner",
	)
	out, err := runCmd.CombinedOutput()
	assert.NoError(t, err, "%s", string(out))

	/*cleanupCmd := exec.Command("docker", "kill", containerName)
	out, err = cleanupCmd.CombinedOutput()
	assert.NoError(t, err, "%s", string(out))
	*/
}

func TestPython(t *testing.T) {
	buildCmd := exec.Command("docker",
		"build",
		"python",
		"-t",
		"python-runner",
	)
	err := buildCmd.Run()
	assert.NoError(t, err)

	beforeSave := `
def before_save(input):
  input["concat"] = input["name"] + " " + str(input["score"])
  return input
`
	customLogic := model.CustomLogic{
		APIID:              "apiID",
		OperationType: model.OperationTypeCreate,
		BeforeSave: &beforeSave,
	}

	customLogicPath, err := writeTmpFile([]model.CustomLogic{customLogic}, "custom-logic-")
	assert.NoError(t, err)

	containerName := uuid.New().String()
	runCmd := exec.Command("docker",
		"run",
		"-d",
		"-v",
		customLogicPath + ":/app/customLogic.json",
		"-p",
		"9090:5000",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		"python-runner",
	)
	out, err := runCmd.CombinedOutput()
	assert.NoError(t, err, "%s", string(out))
}

func writeTmpFile(input interface{}, prefix string) (string, error) {
	file, err := ioutil.TempFile("/Users/gracew/tmp", prefix)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}

	err = json.NewEncoder(file).Encode(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode object to file")
	}
	return filepath.Abs(file.Name())
}
