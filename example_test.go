//go:build unix

package cmd_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/grongor/go-cmd"
	cmd_mocks "github.com/grongor/go-cmd/mocks"
)

// CurrentBranch is an example of code that runs a command; it depends on cmd.Factory instead of os/exec.
func CurrentBranch(factory cmd.Factory) (string, error) {
	output, err := factory.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func ExampleOsExecFactory_Command() {
	command := cmd.NewOsExecFactory().Command("echo", "hello")

	output, err := command.Output()
	if err != nil {
		panic(err)
	}

	fmt.Print(string(output))
	// Output: hello
}

// Example_mocking shows the mocks that replace the real commands in tests. In a test, create them with
// cmd_mocks.NewFactory(t) and cmd_mocks.NewCommand(t) instead, so that their expectations are asserted.
func Example_mocking() {
	factory := &cmd_mocks.Factory{}
	command := &cmd_mocks.Command{}

	factory.EXPECT().Command("git", "branch", "--show-current").Return(command)
	command.EXPECT().Output().Return([]byte("main\n"), nil)

	branch, err := CurrentBranch(factory)

	fmt.Println(branch, err)
	// Output: main <nil>
}

// ExampleExitError shows how to get the exit code and stderr of a failed command.
func ExampleExitError() {
	_, err := cmd.NewOsExecFactory().Command("sh", "-c", "echo oops >&2; exit 3").Output()

	var exitErr cmd.ExitError
	if errors.As(err, &exitErr) {
		fmt.Printf("%d %q\n", exitErr.ExitCode(), exitErr.Stderr())
	}
	// Output: 3 "oops\n"
}

// Example_gracefulTermination terminates a command with SIGTERM when its context is done, and kills it if it
// is still running after the WaitDelay.
func Example_gracefulTermination() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	command := cmd.NewOsExecFactory().CommandContext(ctx, "sleep", "10")
	command.SetCancel(func() error { return command.GetProcess().Signal(syscall.SIGTERM) })
	command.SetWaitDelay(5 * time.Second)

	if err := command.Start(); err != nil {
		panic(err)
	}

	stop() // e.g. on shutdown

	fmt.Println(command.Wait())
	// Output: signal: terminated
}
