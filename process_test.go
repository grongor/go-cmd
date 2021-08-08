package cmd_test

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/grongor/go-cmd"
	"github.com/stretchr/testify/require"
)

func TestExitError(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", ">&2 echo some error ; exit 5")

	output, err := command.Output()
	assert.Empty(output)
	assert.EqualError(err, "exit status 5")

	assert.IsType(&exec.ExitError{}, errors.Unwrap(err))

	var exitErr cmd.ExitError
	assert.True(errors.As(err, &exitErr))
	assert.Equal("some error\n", string(exitErr.Stderr()))

	assert.Equal(5, exitErr.ExitCode())
}

func TestProcess_SignalAndWait(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "5")

	assert.Nil(command.GetProcess())

	assert.NoError(command.Start())

	process := command.GetProcess()
	assert.NotNil(process)

	assert.NoError(process.Signal(os.Interrupt))

	state, err := process.Wait()
	assert.NoError(err)
	assert.False(state.Success())
}

func TestProcess_KillAndRelease(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "5")

	assert.Nil(command.GetProcess())

	assert.NoError(command.Start())

	process := command.GetProcess()
	assert.NotNil(process)
	assert.Same(process, command.GetProcess())

	assert.NoError(process.Kill())
	assert.NoError(process.Release())
}

func TestProcessState(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "0.01")

	assert.Nil(command.GetProcessState())

	assert.NoError(command.Run())

	processState := command.GetProcessState()
	assert.NotNil(processState)

	assert.Equal(0, processState.ExitCode())
	assert.True(processState.Exited())
	assert.Greater(processState.Pid(), 1)
	assert.True(processState.Success())
	assert.Equal(syscall.WaitStatus(0), processState.Sys())
	assert.GreaterOrEqual(processState.SystemTime(), time.Duration(0))
	assert.IsType(&syscall.Rusage{}, processState.SysUsage())
	assert.GreaterOrEqual(processState.UserTime(), time.Duration(0))
}
