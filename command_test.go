package cmd_test

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/grongor/go-cmd"
	"github.com/stretchr/testify/require"
)

func TestOsExecCommand_Run(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	err := command.Run()
	assert.NoError(err)
}

func TestOsExecCommand_StartAndWait(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	buf := &strings.Builder{}

	command.SetStdout(buf)

	err := command.Start()
	assert.NoError(err)

	assert.NoError(command.Wait())

	assert.Equal(buf, command.GetStdout())
	assert.Equal("this is stdout\n", buf.String())
}

func TestOsExecCommand_StartAndWaitWithStderr(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	stdoutBuf := &strings.Builder{}
	stderrBuf := &strings.Builder{}

	command.SetStdout(stdoutBuf)
	command.SetStderr(stderrBuf)

	err := command.Start()
	assert.NoError(err)

	assert.NoError(command.Wait())

	assert.Equal(stdoutBuf, command.GetStdout())
	assert.Equal("this is stdout\n", stdoutBuf.String())
	assert.Equal(stderrBuf, command.GetStderr())
	assert.Equal("and this is stderr\n", stderrBuf.String())
}

func TestOsExecCommand_CombinedOutput(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	output, err := command.CombinedOutput()
	assert.NoError(err)
	assert.Equal("this is stdout\nand this is stderr\n", string(output))
}

func TestOsExecCommand_Output(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", "echo 'this is' stdout ; >&2 echo 'much error' ; exit 123")

	output, err := command.Output()
	assert.EqualError(err, "exit status 123")
	assert.Equal("this is stdout\n", string(output))

	var exitErr cmd.ExitError
	assert.True(errors.As(err, &exitErr))
	assert.Equal("much error\n", string(exitErr.Stderr()))
}

func TestOsExecCommand_Pipes(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", "cat ; >&2 echo 'and stderr'")

	stdin, err := command.StdinPipe()
	assert.NoError(err)

	wg := sync.WaitGroup{}
	wg.Add(2)

	stdout, err := command.StdoutPipe()
	assert.NoError(err)

	stderr, err := command.StderrPipe()
	assert.NoError(err)

	assert.NoError(command.Start())

	stdinContents := "this is from stdin"
	n, err := io.WriteString(stdin, stdinContents)
	assert.Equal(len(stdinContents), n)
	assert.NoError(err)
	assert.NoError(stdin.Close())

	stdoutContent, err := io.ReadAll(stdout)
	assert.NoError(err)

	stderrContent, err := io.ReadAll(stderr)
	assert.NoError(err)

	assert.Equal(stdinContents, string(stdoutContent))
	assert.Equal("and stderr\n", string(stderrContent))

	assert.NoError(command.Wait())
}

func TestOsExecCommand_OtherMethods(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("something", "arguments", "here")

	assert.Equal("something", command.GetPath())
	command.SetPath("changed")
	assert.Equal("changed", command.GetPath())

	assert.Equal([]string{"something", "arguments", "here"}, command.GetArgs())
	command.AppendArgs("another")
	assert.Equal([]string{"something", "arguments", "here", "another"}, command.GetArgs())
	command.SetArgs([]string{"now", "final"})
	assert.Equal([]string{"now", "final"}, command.GetArgs())

	assert.Nil(command.GetEnv())
	command.AppendEnv("LOREM=ipsum", "DOLOR=sit amet")
	assert.Equal([]string{"LOREM=ipsum", "DOLOR=sit amet"}, command.GetEnv())
	command.SetEnv([]string{"WHATEVER=wow"})
	assert.Equal([]string{"WHATEVER=wow"}, command.GetEnv())

	assert.Equal("", command.GetDir())
	command.SetDir("/opt")
	assert.Equal("/opt", command.GetDir())

	reader := strings.NewReader("stdin")
	assert.Nil(command.GetStdin())
	command.SetStdin(reader)
	assert.Same(reader, command.GetStdin())

	file1, err := os.CreateTemp("", "cmd-test")
	assert.NoError(err)
	defer file1.Close()
	defer os.Remove(file1.Name())

	file2, err := os.CreateTemp("", "cmd-test")
	assert.NoError(err)
	defer file2.Close()
	defer os.Remove(file2.Name())

	assert.Nil(command.GetExtraFiles())
	command.AppendExtraFiles(file1)
	assert.Len(command.GetExtraFiles(), 1)
	assert.Same(file1, command.GetExtraFiles()[0])
	command.SetExtraFiles([]*os.File{file2})
	assert.Len(command.GetExtraFiles(), 1)
	assert.Same(file2, command.GetExtraFiles()[0])

	sysProcAttr := &syscall.SysProcAttr{Chroot: "/opt"}
	assert.Nil(command.GetSysProcAttr())
	command.SetSysProcAttr(sysProcAttr)
	assert.Same(sysProcAttr, command.GetSysProcAttr())

	assert.Equal("now final", command.String())
}

func TestOsExecCommand_NotExitError(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("something that doesnt exist")

	err := command.Run()

	var exitErr cmd.ExitError
	assert.False(errors.As(err, &exitErr))
	assert.IsType(&exec.Error{}, err)
}

func TestOsExecCommand_PanicAfterChangingPropertiesAfterStart(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "5")

	assert.NoError(command.Start())

	defer func() {
		command.GetProcess().Kill()
		command.GetProcess().Release()
	}()

	assert.Panics(func() { command.SetPath("whatever") })
	assert.Panics(func() { command.SetArgs([]string{"whatever"}) })
	assert.Panics(func() { command.AppendArgs("whatever") })
	assert.Panics(func() { command.SetEnv([]string{"whatever=lol"}) })
	assert.Panics(func() { command.AppendEnv("whatever") })
	assert.Panics(func() { command.SetStdin(nil) })
	assert.Panics(func() { command.SetStdout(nil) })
	assert.Panics(func() { command.SetStderr(nil) })
	assert.Panics(func() { command.SetExtraFiles(nil) })
	assert.Panics(func() { command.AppendExtraFiles(nil) })
	assert.Panics(func() { command.SetSysProcAttr(nil) })
}

func getCommand() *cmd.OsExecCommand {
	command := "echo 'this is' stdout ; >&2 echo 'and this is' stderr"

	return cmd.NewOsExecFactory().Command("sh", "-c", command).(*cmd.OsExecCommand)
}
