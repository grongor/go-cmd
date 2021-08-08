package cmd

import (
	"context"
	"os/exec"
)

type Factory interface {
	Command(name string, args ...string) Command
	CommandContext(ctx context.Context, name string, args ...string) Command
	LookPath(file string) (string, error)
}

type OsExecFactory struct{}

func (*OsExecFactory) Command(name string, args ...string) Command {
	return NewOsExecCommand(exec.Command(name, args...))
}

func (*OsExecFactory) CommandContext(ctx context.Context, name string, args ...string) Command {
	return NewOsExecCommand(exec.CommandContext(ctx, name, args...))
}

func (*OsExecFactory) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func NewOsExecFactory() *OsExecFactory {
	return &OsExecFactory{}
}
