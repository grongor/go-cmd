package cmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ErrAlreadyStarted is the panic value of the [Command] setters called after the process has been started.
var ErrAlreadyStarted = errors.New("cmd: process already started")

// Command is the interface of [exec.Cmd]. Its fields are replaced by getters and setters.
type Command interface {
	// GetPath returns exec.Cmd.Path.
	GetPath() string
	// SetPath sets exec.Cmd.Path. It panics with ErrAlreadyStarted once the process has been started.
	SetPath(path string)
	// GetArgs returns exec.Cmd.Args.
	GetArgs() []string
	// SetArgs sets exec.Cmd.Args. It panics with ErrAlreadyStarted once the process has been started.
	SetArgs(args []string)
	// AppendArgs appends args to exec.Cmd.Args. It panics with ErrAlreadyStarted once the process has
	// been started.
	AppendArgs(args ...string)
	// GetEnv returns exec.Cmd.Env.
	GetEnv() []string
	// SetEnv sets exec.Cmd.Env. It panics with ErrAlreadyStarted once the process has been started.
	SetEnv(env []string)
	// AppendEnv appends env to exec.Cmd.Env, exactly like cmd.Env = append(cmd.Env, env...). It panics with
	// ErrAlreadyStarted once the process has been started. Note that while Env is nil the command inherits the
	// current process's environment, so appending to a nil Env drops the inherited environment; use
	// SetEnv(append(c.Environ(), env...)) to add variables to it instead.
	AppendEnv(env ...string)
	// GetDir returns exec.Cmd.Dir.
	GetDir() string
	// SetDir sets exec.Cmd.Dir. It panics with ErrAlreadyStarted once the process has been started.
	SetDir(dir string)
	// GetStdin returns exec.Cmd.Stdin.
	GetStdin() io.Reader
	// SetStdin sets exec.Cmd.Stdin. It panics with ErrAlreadyStarted once the process has been started.
	SetStdin(stdin io.Reader)
	// GetStdout returns exec.Cmd.Stdout.
	GetStdout() io.Writer
	// SetStdout sets exec.Cmd.Stdout. It panics with ErrAlreadyStarted once the process has been started.
	SetStdout(stdout io.Writer)
	// GetStderr returns exec.Cmd.Stderr.
	GetStderr() io.Writer
	// SetStderr sets exec.Cmd.Stderr. It panics with ErrAlreadyStarted once the process has been started.
	SetStderr(stderr io.Writer)
	// GetExtraFiles returns exec.Cmd.ExtraFiles.
	GetExtraFiles() []*os.File
	// SetExtraFiles sets exec.Cmd.ExtraFiles. It panics with ErrAlreadyStarted once the process has been started.
	SetExtraFiles(files []*os.File)
	// AppendExtraFiles appends files to exec.Cmd.ExtraFiles. It panics with ErrAlreadyStarted once the process has
	// been started.
	AppendExtraFiles(files ...*os.File)
	// GetSysProcAttr returns exec.Cmd.SysProcAttr.
	GetSysProcAttr() *syscall.SysProcAttr
	// SetSysProcAttr sets exec.Cmd.SysProcAttr. It panics with ErrAlreadyStarted once the process has been started.
	SetSysProcAttr(attr *syscall.SysProcAttr)
	// GetCancel returns exec.Cmd.Cancel.
	GetCancel() func() error
	// SetCancel sets exec.Cmd.Cancel. It panics with ErrAlreadyStarted once the process has been started.
	SetCancel(cancel func() error)
	// GetWaitDelay returns exec.Cmd.WaitDelay.
	GetWaitDelay() time.Duration
	// SetWaitDelay sets exec.Cmd.WaitDelay. It panics with ErrAlreadyStarted once the process has been started.
	SetWaitDelay(delay time.Duration)
	// GetErr returns exec.Cmd.Err.
	GetErr() error
	// SetErr sets exec.Cmd.Err. It panics with ErrAlreadyStarted once the process has been started.
	SetErr(err error)
	// GetProcess returns exec.Cmd.Process wrapped in a Process, or nil if the process hasn't been started.
	GetProcess() Process
	// GetProcessState returns exec.Cmd.ProcessState wrapped in a ProcessState, or nil if the process hasn't
	// been waited for.
	GetProcessState() ProcessState

	// CombinedOutput is [exec.Cmd.CombinedOutput].
	CombinedOutput() ([]byte, error)
	// Environ is [exec.Cmd.Environ].
	Environ() []string
	// Output is [exec.Cmd.Output].
	Output() ([]byte, error)
	// Run is [exec.Cmd.Run].
	Run() error
	// Start is [exec.Cmd.Start].
	Start() error
	// StdinPipe is [exec.Cmd.StdinPipe].
	StdinPipe() (io.WriteCloser, error)
	// StdoutPipe is [exec.Cmd.StdoutPipe].
	StdoutPipe() (io.ReadCloser, error)
	// StderrPipe is [exec.Cmd.StderrPipe].
	StderrPipe() (io.ReadCloser, error)
	// String is [exec.Cmd.String].
	String() string
	// Wait is [exec.Cmd.Wait].
	Wait() error
}

var _ Command = (*OsExecCommand)(nil)

// OsExecCommand implements [Command] by wrapping an *exec.Cmd; every method behaves exactly like its exec.Cmd
// counterpart. Like exec.Cmd, it is not safe for concurrent use, except for what exec.Cmd itself allows (such as
// reading the process from a Cancel function).
type OsExecCommand struct {
	command *exec.Cmd
}

// Cmd returns the wrapped *exec.Cmd, for what the [Command] interface doesn't cover (e.g. libraries that take an
// *exec.Cmd). It isn't part of the interface, so code using it can't be tested with the mocks.
func (c *OsExecCommand) Cmd() *exec.Cmd {
	return c.command
}

// GetPath returns exec.Cmd.Path.
func (c *OsExecCommand) GetPath() string {
	return c.command.Path
}

// SetPath sets exec.Cmd.Path. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetPath(path string) {
	c.checkAlreadyStarted()

	c.command.Path = path
}

// GetArgs returns exec.Cmd.Args.
func (c *OsExecCommand) GetArgs() []string {
	return c.command.Args
}

// SetArgs sets exec.Cmd.Args. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetArgs(args []string) {
	c.checkAlreadyStarted()

	c.command.Args = args
}

// AppendArgs appends args to exec.Cmd.Args. It panics with [ErrAlreadyStarted]
// once the process has been started.
func (c *OsExecCommand) AppendArgs(args ...string) {
	c.checkAlreadyStarted()

	c.command.Args = append(c.command.Args, args...)
}

// GetEnv returns exec.Cmd.Env.
func (c *OsExecCommand) GetEnv() []string {
	return c.command.Env
}

// SetEnv sets exec.Cmd.Env. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetEnv(env []string) {
	c.checkAlreadyStarted()

	c.command.Env = env
}

// AppendEnv appends env to exec.Cmd.Env, exactly like cmd.Env = append(cmd.Env, env...). It panics with
// [ErrAlreadyStarted] once the process has been started. Note that while Env is nil the command inherits the
// current process's environment, so appending to a nil Env drops the inherited environment; use
// SetEnv(append(c.Environ(), env...)) to add variables to it instead.
func (c *OsExecCommand) AppendEnv(env ...string) {
	c.checkAlreadyStarted()

	c.command.Env = append(c.command.Env, env...)
}

// GetDir returns exec.Cmd.Dir.
func (c *OsExecCommand) GetDir() string {
	return c.command.Dir
}

// SetDir sets exec.Cmd.Dir. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetDir(dir string) {
	c.checkAlreadyStarted()

	c.command.Dir = dir
}

// GetStdin returns exec.Cmd.Stdin.
func (c *OsExecCommand) GetStdin() io.Reader {
	return c.command.Stdin
}

// SetStdin sets exec.Cmd.Stdin. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetStdin(stdin io.Reader) {
	c.checkAlreadyStarted()

	c.command.Stdin = stdin
}

// GetStdout returns exec.Cmd.Stdout.
func (c *OsExecCommand) GetStdout() io.Writer {
	return c.command.Stdout
}

// SetStdout sets exec.Cmd.Stdout. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetStdout(stdout io.Writer) {
	c.checkAlreadyStarted()

	c.command.Stdout = stdout
}

// GetStderr returns exec.Cmd.Stderr.
func (c *OsExecCommand) GetStderr() io.Writer {
	return c.command.Stderr
}

// SetStderr sets exec.Cmd.Stderr. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetStderr(stderr io.Writer) {
	c.checkAlreadyStarted()

	c.command.Stderr = stderr
}

// GetExtraFiles returns exec.Cmd.ExtraFiles.
func (c *OsExecCommand) GetExtraFiles() []*os.File {
	return c.command.ExtraFiles
}

// SetExtraFiles sets exec.Cmd.ExtraFiles. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetExtraFiles(files []*os.File) {
	c.checkAlreadyStarted()

	c.command.ExtraFiles = files
}

// AppendExtraFiles appends files to exec.Cmd.ExtraFiles. It panics with [ErrAlreadyStarted]
// once the process has been started.
func (c *OsExecCommand) AppendExtraFiles(files ...*os.File) {
	c.checkAlreadyStarted()

	c.command.ExtraFiles = append(c.command.ExtraFiles, files...)
}

// GetSysProcAttr returns exec.Cmd.SysProcAttr.
func (c *OsExecCommand) GetSysProcAttr() *syscall.SysProcAttr {
	return c.command.SysProcAttr
}

// SetSysProcAttr sets exec.Cmd.SysProcAttr. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetSysProcAttr(attr *syscall.SysProcAttr) {
	c.checkAlreadyStarted()

	c.command.SysProcAttr = attr
}

// GetCancel returns exec.Cmd.Cancel.
func (c *OsExecCommand) GetCancel() func() error {
	return c.command.Cancel
}

// SetCancel sets exec.Cmd.Cancel. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetCancel(cancel func() error) {
	c.checkAlreadyStarted()

	c.command.Cancel = cancel
}

// GetWaitDelay returns exec.Cmd.WaitDelay.
func (c *OsExecCommand) GetWaitDelay() time.Duration {
	return c.command.WaitDelay
}

// SetWaitDelay sets exec.Cmd.WaitDelay. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetWaitDelay(delay time.Duration) {
	c.checkAlreadyStarted()

	c.command.WaitDelay = delay
}

// GetErr returns exec.Cmd.Err.
func (c *OsExecCommand) GetErr() error {
	return c.command.Err
}

// SetErr sets exec.Cmd.Err. It panics with [ErrAlreadyStarted] once the process has been started.
func (c *OsExecCommand) SetErr(err error) {
	c.checkAlreadyStarted()

	c.command.Err = err
}

// GetProcess returns exec.Cmd.Process wrapped in an [OsProcess], or nil if the process hasn't been started.
func (c *OsExecCommand) GetProcess() Process {
	if c.command.Process == nil {
		return nil
	}

	return NewOsProcess(c.command.Process)
}

// GetProcessState returns exec.Cmd.ProcessState wrapped in an [OsProcessState], or nil if the process hasn't
// been waited for.
func (c *OsExecCommand) GetProcessState() ProcessState {
	if c.command.ProcessState == nil {
		return nil
	}

	return NewOsProcessState(c.command.ProcessState)
}

// CombinedOutput calls [exec.Cmd.CombinedOutput]; an *exec.ExitError is returned as an [*OsExecExitError].
func (c *OsExecCommand) CombinedOutput() ([]byte, error) {
	output, err := c.command.CombinedOutput()

	return output, wrapErr(err)
}

// Environ calls [exec.Cmd.Environ].
func (c *OsExecCommand) Environ() []string {
	return c.command.Environ()
}

// Output calls [exec.Cmd.Output]; an *exec.ExitError is returned as an [*OsExecExitError].
func (c *OsExecCommand) Output() ([]byte, error) {
	output, err := c.command.Output()

	return output, wrapErr(err)
}

// Run calls [exec.Cmd.Run]; an *exec.ExitError is returned as an [*OsExecExitError].
func (c *OsExecCommand) Run() error {
	return wrapErr(c.command.Run())
}

// Start calls [exec.Cmd.Start].
func (c *OsExecCommand) Start() error {
	return c.command.Start()
}

// StdinPipe calls [exec.Cmd.StdinPipe].
func (c *OsExecCommand) StdinPipe() (io.WriteCloser, error) {
	return c.command.StdinPipe()
}

// StdoutPipe calls [exec.Cmd.StdoutPipe].
func (c *OsExecCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.command.StdoutPipe()
}

// StderrPipe calls [exec.Cmd.StderrPipe].
func (c *OsExecCommand) StderrPipe() (io.ReadCloser, error) {
	return c.command.StderrPipe()
}

// String calls [exec.Cmd.String].
func (c *OsExecCommand) String() string {
	return c.command.String()
}

// Wait calls [exec.Cmd.Wait]; an *exec.ExitError is returned as an [*OsExecExitError].
func (c *OsExecCommand) Wait() error {
	return wrapErr(c.command.Wait())
}

func (c *OsExecCommand) checkAlreadyStarted() {
	if c.command.Process != nil {
		panic(ErrAlreadyStarted)
	}
}

// wrapErr converts an *exec.ExitError returned by exec.Cmd. Only a bare *exec.ExitError is converted: os/exec
// returns it unwrapped, and any other error (even one wrapping an *exec.ExitError, e.g. from a Cancel function)
// must be returned unchanged.
func wrapErr(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok { //nolint:errorlint // only the bare error, see above
		return NewOsExecExitError(exitErr)
	}

	return err
}

// NewOsExecCommand wraps command in an [OsExecCommand].
func NewOsExecCommand(command *exec.Cmd) *OsExecCommand {
	return &OsExecCommand{command: command}
}
