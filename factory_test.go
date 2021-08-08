package cmd_test

import (
	"context"
	"os"
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

	wd, err := os.Getwd()
	assert.NoError(err)

	os.Setenv("PATH", wd+string(os.PathSeparator)+"bin"+string(os.PathListSeparator)+os.Getenv("PATH"))

	path, err := factory.LookPath("mockery")
	assert.Equal(wd+"/bin/mockery", path)
	assert.NoError(err)
}
