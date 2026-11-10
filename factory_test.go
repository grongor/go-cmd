//go:build unix

package cmd_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grongor/go-cmd"
	"github.com/stretchr/testify/require"
)

func TestOsExecFactory_Command(t *testing.T) {
	assert := require.New(t)

	factory := cmd.NewOsExecFactory()

	command := factory.Command("echo", "something to", "output")

	output, err := command.CombinedOutput()
	assert.NoError(err)

	assert.Equal("something to output\n", string(output))
}

func TestOsExecFactory_CommandContext(t *testing.T) {
	assert := require.New(t)

	factory := cmd.NewOsExecFactory()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	command := factory.CommandContext(ctx, "sleep", "1")

	output, err := command.CombinedOutput()
	assert.Empty(output)
	assert.Error(err)
	assert.Contains([]string{"context deadline exceeded", "signal: killed"}, err.Error())
	assert.Equal(context.DeadlineExceeded, ctx.Err())
}

func TestOsExecFactory_LookupPath(t *testing.T) {
	assert := require.New(t)

	factory := cmd.NewOsExecFactory()

	dir := t.TempDir()
	executable := filepath.Join(dir, "go-cmd-test-executable")

	assert.NoError(os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o700)) //nolint:gosec // must be executable

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	path, err := factory.LookPath("go-cmd-test-executable")
	assert.NoError(err)
	assert.Equal(executable, path)
}
