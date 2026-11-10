package cmd

import (
	"os"
	"os/exec"
	"time"
)

// ExitError is the interface of [exec.ExitError]: the [ProcessState] of a process that exited unsuccessfully,
// plus the stderr output collected by [Command.Output].
type ExitError interface {
	error
	ProcessState

	// Stderr is exec.ExitError.Stderr.
	Stderr() []byte
}

var _ ExitError = (*OsExecExitError)(nil)

// OsExecExitError implements [ExitError] by wrapping an *exec.ExitError. errors.As and errors.Is see through it
// to the *exec.ExitError.
type OsExecExitError struct {
	*exec.ExitError
}

// Stderr returns exec.ExitError.Stderr.
func (e *OsExecExitError) Stderr() []byte {
	return e.ExitError.Stderr
}

// Unwrap returns the wrapped *exec.ExitError.
func (e *OsExecExitError) Unwrap() error {
	return e.ExitError
}

// NewOsExecExitError wraps err in an [OsExecExitError].
func NewOsExecExitError(err *exec.ExitError) *OsExecExitError {
	return &OsExecExitError{ExitError: err}
}

// Process is the interface of [os.Process].
type Process interface {
	// Pid is os.Process.Pid.
	Pid() int
	// Kill is [os.Process.Kill].
	Kill() error
	// Release is [os.Process.Release].
	Release() error
	// Signal is [os.Process.Signal].
	Signal(sig os.Signal) error
	// Wait is [os.Process.Wait].
	Wait() (ProcessState, error)
	// WithHandle is [os.Process.WithHandle].
	WithHandle(f func(handle uintptr)) error
}

var _ Process = (*OsProcess)(nil)

// OsProcess implements [Process] by wrapping an *os.Process.
type OsProcess struct {
	*os.Process
}

// Pid returns os.Process.Pid.
func (p *OsProcess) Pid() int {
	return p.Process.Pid
}

// Wait calls [os.Process.Wait] and wraps the returned state in an [OsProcessState] (nil if there is none).
func (p *OsProcess) Wait() (ProcessState, error) {
	state, err := p.Process.Wait()
	if state != nil {
		return NewOsProcessState(state), err
	}

	return nil, err
}

// NewOsProcess wraps process in an [OsProcess].
func NewOsProcess(process *os.Process) *OsProcess {
	return &OsProcess{Process: process}
}

// ProcessState is the interface of [os.ProcessState].
type ProcessState interface {
	// ExitCode is [os.ProcessState.ExitCode].
	ExitCode() int
	// Exited is [os.ProcessState.Exited].
	Exited() bool
	// Pid is [os.ProcessState.Pid].
	Pid() int
	// String is [os.ProcessState.String].
	String() string
	// Success is [os.ProcessState.Success].
	Success() bool
	// Sys is [os.ProcessState.Sys].
	Sys() any
	// SystemTime is [os.ProcessState.SystemTime].
	SystemTime() time.Duration
	// SysUsage is [os.ProcessState.SysUsage].
	SysUsage() any
	// UserTime is [os.ProcessState.UserTime].
	UserTime() time.Duration
}

var _ ProcessState = (*OsProcessState)(nil)

// OsProcessState implements [ProcessState] by wrapping an *os.ProcessState.
type OsProcessState struct {
	*os.ProcessState
}

// NewOsProcessState wraps state in an [OsProcessState].
func NewOsProcessState(state *os.ProcessState) *OsProcessState {
	return &OsProcessState{ProcessState: state}
}
