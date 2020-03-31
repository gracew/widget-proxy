package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fileContent struct {
	filename string
	content  string
}

type testResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func TestNode(t *testing.T) {
	beforeCreate := `
function beforeCreate(input) {
	input.message = "Hello " + input.name;
	return input;
}

module.exports = beforeCreate;
`
	afterCreate := `
function afterCreate(input) {
	input.message = "Bye " + input.name;
	return input;
}

module.exports = afterCreate;
`

	localPort := "7070"
	container := setupContainer(t, "node", localPort, []fileContent{
		fileContent{filename: "beforecreate.js", content: beforeCreate},
		fileContent{filename: "aftercreate.js", content: afterCreate},
	})
	defer tearDownContainer(t, container)

	res1, err := http.Post(fmt.Sprintf("http://localhost:%s/beforecreate", localPort), "application/json", strings.NewReader(`{"name": "Jane"}`))
	assert.NoError(t, err)
	assert.Equal(t, testResponse{Name: "Jane", Message: "Hello Jane"}, decode(t, res1.Body))

	res2, err := http.Post(fmt.Sprintf("http://localhost:%s/aftercreate", localPort), "application/json", strings.NewReader(`{"name": "Jane"}`))
	assert.NoError(t, err)
	assert.Equal(t, testResponse{Name: "Jane", Message: "Bye Jane"}, decode(t, res2.Body))
}

func TestPython(t *testing.T) {
	beforeCreate := `
def beforeCreate(input):
  input["message"] = "Hello " + input["name"]
  return input
`
	afterCreate := `
def afterCreate(input):
  input["message"] = "Bye " + input["name"]
  return input
`
	localPort := "7071"
	container := setupContainer(t, "python", localPort, []fileContent{
		fileContent{filename: "beforecreate.py", content: beforeCreate},
		fileContent{filename: "aftercreate.py", content: afterCreate},
	})
	defer tearDownContainer(t, container)

	res1, err := http.Post(fmt.Sprintf("http://localhost:%s/beforecreate", localPort), "application/json", strings.NewReader(`{"name": "Jane"}`))
	assert.NoError(t, err)
	assert.Equal(t, testResponse{Name: "Jane", Message: "Hello Jane"}, decode(t, res1.Body))

	res2, err := http.Post(fmt.Sprintf("http://localhost:%s/aftercreate", localPort), "application/json", strings.NewReader(`{"name": "Jane"}`))
	assert.NoError(t, err)
	assert.Equal(t, testResponse{Name: "Jane", Message: "Bye Jane"}, decode(t, res2.Body))
}

// returns the container name
func setupContainer(t *testing.T, buildPath string, localPort string, fileContents []fileContent) string {
	imageName := buildPath + "-runner"
	buildCmd := exec.Command("docker",
		"build",
		buildPath,
		"-t",
		imageName,
	)
	out, err := buildCmd.CombinedOutput()
	t.Log(string(out))
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
		localPort+":8080",
		"--name",
		containerName,
		"--network",
		"widget-proxy_default",
		imageName,
	)
	out, err = runCmd.CombinedOutput()
	t.Log(string(out))
	assert.NoError(t, err)

	waitForServer(t, localPort)
	return containerName
}

func writeFileInDir(dir string, name string, input string) error {
	path := filepath.Join(dir, name)
	err := ioutil.WriteFile(path, []byte(input), 0644)
	if err != nil {
		return err
	}
	return nil
}

// waits up to one second for server to start
func waitForServer(t *testing.T, localPort string) {
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		res, err := http.Get(fmt.Sprintf("http://localhost:%s/ping", localPort))
		if err == nil && res.StatusCode == 200 {
			return
		}
	}
	assert.Fail(t, "server failed to start within 1 second")
}

func decode(t *testing.T, body io.Reader) testResponse {
	var res testResponse
	err := json.NewDecoder(body).Decode(&res)
	assert.NoError(t, err)
	return res
}

func tearDownContainer(t *testing.T, containerName string) {
	cmd := exec.Command("docker",
		"stop",
		// "-f",
		containerName,
	)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	assert.NoError(t, err)
}
