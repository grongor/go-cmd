package cmd

import (
	"os"
	"os/exec"
	"time"
)

type ExitError interface {
	error
	ProcessState

	Stderr() []byte
}

type OsExecExitError struct {
	*exec.ExitError
}

func (e *OsExecExitError) Stderr() []byte {
	return e.ExitError.Stderr
}

func (e *OsExecExitError) Unwrap() error {
	return e.ExitError
}

func NewOsExecExitError(err *exec.ExitError) *OsExecExitError {
	return &OsExecExitError{ExitError: err}
}

type Process interface {
	Kill() error
	Release() error
	Signal(sig os.Signal) error
	Wait() (ProcessState, error)
}

type OsProcess struct {
	*os.Process
}

func (p *OsProcess) Wait() (ProcessState, error) {
	state, err := p.Process.Wait()
	if state != nil {
		return NewOsProcessState(state), err
	}

	return nil, err
}

func NewOsProcess(process *os.Process) *OsProcess {
	return &OsProcess{Process: process}
}

type ProcessState interface {
	ExitCode() int
	Exited() bool
	Pid() int
	Success() bool
	Sys() interface{}
	SystemTime() time.Duration
	SysUsage() interface{}
	UserTime() time.Duration
}

type OsProcessState struct {
	*os.ProcessState
}

func NewOsProcessState(state *os.ProcessState) *OsProcessState {
	return &OsProcessState{ProcessState: state}
}
