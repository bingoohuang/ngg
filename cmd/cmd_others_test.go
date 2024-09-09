//go:build !windows

package cmd_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/cmd"
	"github.com/stretchr/testify/assert"
)

func TestCommand_ExecuteStderr1(t *testing.T) {
	c := cmd.New(">&2 echo hello")
	err := c.Run(context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, "hello\n", c.Stderr())
}

func TestCommand_WithTimeout1(t *testing.T) {
	c := cmd.New("sleep 0.1;", cmd.WithTimeout(1*time.Millisecond))
	err := c.Run(context.TODO())

	assert.NotNil(t, err)
	// Sadly a process can not be killed every time :(
	containsMsg := strings.Contains(err.Error(), "timeout, kill") ||
		strings.Contains(err.Error(), "timeout 1ms")
	assert.True(t, containsMsg)
}

func TestCommand_WithValidTimeout1(t *testing.T) {
	c := cmd.New("sleep 0.01;", cmd.WithTimeout(500*time.Millisecond))
	err := c.Run(context.TODO())
	assert.Nil(t, err)
}

func TestCommand_WithWorkingDir(t *testing.T) {
	tempDir := os.TempDir()
	setWorkingDir := func(c *cmd.Cmd) {
		c.WorkingDir = tempDir
	}

	c := cmd.New("pwd", setWorkingDir)
	c.Run(context.TODO())

	out := c.Stdout()
	assert.True(t, strings.Contains(out, tempDir[:len(tempDir)-1]))
}

func TestCommand_WithStandardStreams(t *testing.T) {
	tmpFile, _ := os.CreateTemp("/tmp", "stdout_")
	originalStdout := os.Stdout
	os.Stdout = tmpFile

	// Reset os.Stdout to its original value
	defer func() {
		os.Stdout = originalStdout
	}()

	c := cmd.New("echo hey", cmd.WithStdStreams())
	c.Run(context.TODO())

	r, err := os.ReadFile(tmpFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "hey\n", string(r))
}

func TestCommand_WithoutTimeout(t *testing.T) {
	c := cmd.New("sleep 0.001; echo hello", cmd.WithTimeout(0))
	err := c.Run(context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, "hello\n", c.Stdout())
}

func TestCommand_WithInvalidDir(t *testing.T) {
	c := cmd.New("echo hello", cmd.WithWorkingDir("/invalid"))
	err := c.Run(context.TODO())
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), ": no such file or directory"))
}

func TestWithInheritedEnvironment(t *testing.T) {
	os.Setenv("FROM_OS", "is on os")
	os.Setenv("OVERWRITE", "is on os but should be overwritten")
	defer func() {
		os.Unsetenv("FROM_OS")
		os.Unsetenv("OVERWRITE")
	}()

	c := cmd.New(
		"echo $FROM_OS $OVERWRITE",
		cmd.AddEnv(map[string]string{"OVERWRITE": "overwritten"}))
	c.Run(context.TODO())

	assertEqualWithLineBreak(t, "is on os overwritten", c.Stdout())
}

func TestWithCustomStderr(t *testing.T) {
	writer := bytes.Buffer{}
	c := cmd.New(">&2 echo StderrBuf; sleep 0.01; echo StdoutBuf;", cmd.WithStderr(&writer))
	c.Run(context.TODO())

	assertEqualWithLineBreak(t, "StderrBuf", writer.String())
	assertEqualWithLineBreak(t, "StdoutBuf", c.Stdout())
	assertEqualWithLineBreak(t, "StderrBuf", c.Stderr())
	assertEqualWithLineBreak(t, "StderrBuf\nStdoutBuf", c.Combined())
}

func TestWithCustomStdout(t *testing.T) {
	writer := bytes.Buffer{}
	c := cmd.New(">&2 echo StderrBuf; sleep 0.01; echo StdoutBuf;", cmd.WithStdout(&writer))
	c.Run(context.TODO())

	assertEqualWithLineBreak(t, "StdoutBuf", writer.String())
	assertEqualWithLineBreak(t, "StdoutBuf", c.Stdout())
	assertEqualWithLineBreak(t, "StderrBuf", c.Stderr())
	assertEqualWithLineBreak(t, "StderrBuf\nStdoutBuf", c.Combined())
}

func TestWithEnvironmentVariables(t *testing.T) {
	c := cmd.New("echo $Env", cmd.AddEnv(map[string]string{"Env": "value"}))
	c.Run(context.TODO())

	assertEqualWithLineBreak(t, "value", c.Stdout())
}

func TestCommand_WithContext(t *testing.T) {
	// ensure legacy timeout is honored
	c := cmd.New("sleep 3;", cmd.WithTimeout(1*time.Second))
	err := c.Run(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, "timeout 1s: timeout", err.Error())

	// set context timeout to 2 seconds to ensure
	// context takes precedence over timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c = cmd.New("sleep 3;", cmd.WithTimeout(1*time.Second))
	err = c.Run(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "context deadline exceeded", err.Error())
}

func TestCommand_WithCmd(t *testing.T) {
	c := cmd.New(
		"echo $0",
		cmd.WithCmd(exec.Command("/bin/bash", "-c")),
	)

	err := c.Run(context.TODO())
	assert.Nil(t, err)
	// on darwin we use /bin/sh by default test if we're using bash
	assert.NotEqual(t, "/bin/sh\n", c.Stdout())
	assert.Equal(t, "/bin/bash\n", c.Stdout())
}

func TestCommand_ExecuteStderr(t *testing.T) {
	c := cmd.New(">&2 echo hello")
	err := c.Run(context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, "hello\n", c.Stderr())
}

func TestCommand_WithTimeout(t *testing.T) {
	c := cmd.New("sleep 0.5;", cmd.WithTimeout(5*time.Millisecond))
	err := c.Run(context.TODO())

	assert.NotNil(t, err)
	assert.Equal(t, "timeout 5ms: timeout", err.Error())
}

func TestCommand_WithValidTimeout(t *testing.T) {
	c := cmd.New("sleep 0.01;", cmd.WithTimeout(500*time.Millisecond))
	err := c.Run(context.TODO())

	assert.Nil(t, err)
}
