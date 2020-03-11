package docker

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
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

	beforeCreate := `
function beforeSave(input) {
	input.concat = input.name + " " + input.score;
	return input;
}

module.exports = beforeSave;
`

	customLogicDir, err := ioutil.TempDir("/Users/gracew/tmp", "customLogic-")
	assert.NoError(t, err)

	err = writeFileInDir(customLogicDir, "beforeCreate.js", beforeCreate)
	assert.NoError(t, err)

	containerName := uuid.New().String()
	runCmd := exec.Command("docker",
		"run",
		"-d",
		"-v",
		customLogicDir + ":/app/customLogic",
		"-p",
		"7070:8080",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		"node-runner",
	)
	out, err := runCmd.CombinedOutput()
	assert.NoError(t, err, "%s", string(out))
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
def beforeCreate(input):
  input["concat"] = input["name"] + " " + str(input["score"])
  return input
`
	customLogicDir, err := ioutil.TempDir("/Users/gracew/tmp", "customLogic-")
	assert.NoError(t, err)

	err = writeFileInDir(customLogicDir, "beforeCreate.py", beforeSave)
	assert.NoError(t, err)

	containerName := uuid.New().String()
	runCmd := exec.Command("docker",
		"run",
		"-d",
		"-v",
		customLogicDir + ":/app/customLogic",
		"-p",
		"7070:8080",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		"python-runner",
	)
	out, err := runCmd.CombinedOutput()
	assert.NoError(t, err, "%s", string(out))
}

func writeFileInDir(dir string, name string, input string) error {
	path := filepath.Join(dir, name)
	err := ioutil.WriteFile(path, []byte(input), 0644)
	if err != nil {
		return err
	}
	return nil
}