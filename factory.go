package cmd

import (
	"context"
	"os/exec"
)

// Factory creates commands; it replaces the os/exec package-level functions.
type Factory interface {
	// Command is [exec.Command].
	Command(name string, args ...string) Command
	// CommandContext is [exec.CommandContext].
	CommandContext(ctx context.Context, name string, args ...string) Command
	// LookPath is [exec.LookPath].
	LookPath(file string) (string, error)
}

var _ Factory = OsExecFactory{}

// OsExecFactory implements [Factory] using the os/exec package. Its zero value is ready to use.
type OsExecFactory struct{}

// Command returns exec.Command(name, args...) wrapped in an [OsExecCommand].
func (OsExecFactory) Command(name string, args ...string) Command {
	return NewOsExecCommand(exec.Command(name, args...))
}

// CommandContext returns exec.CommandContext(ctx, name, args...) wrapped in an [OsExecCommand].
func (OsExecFactory) CommandContext(ctx context.Context, name string, args ...string) Command {
	return NewOsExecCommand(exec.CommandContext(ctx, name, args...))
}

// LookPath calls [exec.LookPath].
func (OsExecFactory) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// NewOsExecFactory returns an [OsExecFactory].
func NewOsExecFactory() *OsExecFactory {
	return &OsExecFactory{}
}
