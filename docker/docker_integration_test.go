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
	beforeCreate := `
function beforeSave(input) {
	input.concat = input.name + " " + input.score;
	return input;
}

module.exports = beforeSave;
`

	setupContainer(t, "node-runner", []fileContent{fileContent{filename: "beforecreate.js", content: beforeCreate}})
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
		customLogicDir+":/app/customLogic",
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

type fileContent struct {
	filename string
	content string
}

// returns the container name
func setupContainer(t *testing.T, buildPath string, fileContents []fileContent) string {
	buildCmd := exec.Command("docker",
		"build",
		buildPath,
		"-t",
		buildPath + "-runner",
	)
	err := buildCmd.Run()
	assert.NoError(t, err)

	customLogicDir, err := ioutil.TempDir("/Users/gracew/tmp", "customLogic-")
	assert.NoError(t, err)

	for _, fileContent := range fileContents {
		err = writeFileInDir(customLogicDir, fileContent.filename, fileContent.content)
		assert.NoError(t, err)
	}

	containerName := uuid.New().String()
	runCmd := exec.Command("docker",
		"run",
		"-d",
		"-v",
		customLogicDir+":/app/customLogic",
		"-p",
		"7070:8080",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		buildPath,
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
