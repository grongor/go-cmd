package cmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
)

var ErrAlreadyStarted = errors.New("cmd: process already started")

type Command interface {
	// Methods that replace direct access to the exec.Cmd struct properties

	GetPath() string
	SetPath(path string)
	GetArgs() []string
	SetArgs(args []string)
	AppendArgs(args ...string)
	GetEnv() []string
	SetEnv(env []string)
	AppendEnv(env ...string)
	GetDir() string
	SetDir(dir string)
	GetStdin() io.Reader
	SetStdin(reader io.Reader)
	GetStdout() io.Writer
	SetStdout(writer io.Writer)
	GetStderr() io.Writer
	SetStderr(writer io.Writer)
	GetExtraFiles() []*os.File
	SetExtraFiles(files []*os.File)
	AppendExtraFiles(files ...*os.File)
	GetSysProcAttr() *syscall.SysProcAttr
	SetSysProcAttr(attributes *syscall.SysProcAttr)
	GetProcess() Process
	GetProcessState() ProcessState

	// Wrapped methods of the exec.Cmd struct

	CombinedOutput() ([]byte, error)
	Output() ([]byte, error)
	Run() error
	Start() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	String() string
	Wait() error
}

type OsExecCommand struct {
	command *exec.Cmd
	process Process
}

func (c *OsExecCommand) GetPath() string {
	return c.command.Path
}

func (c *OsExecCommand) SetPath(path string) {
	c.checkAlreadyStarted()

	c.command.Path = path
}

func (c *OsExecCommand) GetArgs() []string {
	return c.command.Args
}

func (c *OsExecCommand) SetArgs(args []string) {
	c.checkAlreadyStarted()

	c.command.Args = args
}

func (c *OsExecCommand) AppendArgs(args ...string) {
	c.checkAlreadyStarted()

	c.command.Args = append(c.command.Args, args...)
}

func (c *OsExecCommand) GetEnv() []string {
	return c.command.Env
}

func (c *OsExecCommand) SetEnv(env []string) {
	c.checkAlreadyStarted()

	c.command.Env = env
}

func (c *OsExecCommand) AppendEnv(args ...string) {
	c.checkAlreadyStarted()

	c.command.Env = append(c.command.Env, args...)
}

func (c *OsExecCommand) GetDir() string {
	return c.command.Dir
}

func (c *OsExecCommand) SetDir(dir string) {
	c.checkAlreadyStarted()

	c.command.Dir = dir
}

func (c *OsExecCommand) GetStdin() io.Reader {
	return c.command.Stdin
}

func (c *OsExecCommand) SetStdin(reader io.Reader) {
	c.checkAlreadyStarted()

	c.command.Stdin = reader
}

func (c *OsExecCommand) GetStdout() io.Writer {
	return c.command.Stdout
}

func (c *OsExecCommand) SetStdout(writer io.Writer) {
	c.checkAlreadyStarted()

	c.command.Stdout = writer
}

func (c *OsExecCommand) GetStderr() io.Writer {
	return c.command.Stderr
}

func (c *OsExecCommand) SetStderr(writer io.Writer) {
	c.checkAlreadyStarted()

	c.command.Stderr = writer
}

func (c *OsExecCommand) GetExtraFiles() []*os.File {
	return c.command.ExtraFiles
}

func (c *OsExecCommand) SetExtraFiles(files []*os.File) {
	c.checkAlreadyStarted()

	c.command.ExtraFiles = files
}

func (c *OsExecCommand) AppendExtraFiles(files ...*os.File) {
	c.checkAlreadyStarted()

	c.command.ExtraFiles = append(c.command.ExtraFiles, files...)
}

func (c *OsExecCommand) GetSysProcAttr() *syscall.SysProcAttr {
	return c.command.SysProcAttr
}

func (c *OsExecCommand) SetSysProcAttr(sysProcAttr *syscall.SysProcAttr) {
	c.checkAlreadyStarted()

	c.command.SysProcAttr = sysProcAttr
}

func (c *OsExecCommand) GetProcess() Process {
	if c.process != nil {
		return c.process
	}

	if c.command.Process != nil {
		c.process = NewOsProcess(c.command.Process)
	}

	return c.process
}

func (c *OsExecCommand) GetProcessState() ProcessState {
	return c.command.ProcessState
}

func (c *OsExecCommand) CombinedOutput() ([]byte, error) {
	output, err := c.command.CombinedOutput()

	return output, c.wrapErr(err)
}

func (c *OsExecCommand) Output() ([]byte, error) {
	output, err := c.command.Output()

	return output, c.wrapErr(err)
}

func (c *OsExecCommand) Run() error {
	return c.wrapErr(c.command.Run())
}

func (c *OsExecCommand) Start() error {
	return c.command.Start()
}

func (c *OsExecCommand) StdinPipe() (io.WriteCloser, error) {
	return c.command.StdinPipe()
}

func (c *OsExecCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.command.StdoutPipe()
}

func (c *OsExecCommand) StderrPipe() (io.ReadCloser, error) {
	return c.command.StderrPipe()
}

func (c *OsExecCommand) String() string {
	return c.command.String()
}

func (c *OsExecCommand) Wait() error {
	return c.wrapErr(c.command.Wait())
}

func (*OsExecCommand) wrapErr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return NewOsExecExitError(exitErr)
	}

	return err
}

func (c *OsExecCommand) checkAlreadyStarted() {
	if c.command.Process != nil {
		panic(ErrAlreadyStarted)
	}
}

func NewOsExecCommand(command *exec.Cmd) *OsExecCommand {
	return &OsExecCommand{command: command}
}
