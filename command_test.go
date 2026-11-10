//go:build unix

package cmd_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/grongor/go-cmd"
	"github.com/stretchr/testify/require"
)

func TestOsExecCommand_Run(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	err := command.Run()
	assert.NoError(err)
}

func TestOsExecCommand_StartAndWait(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	buf := &strings.Builder{}

	command.SetStdout(buf)

	err := command.Start()
	assert.NoError(err)

	assert.NoError(command.Wait())

	assert.Same(buf, command.GetStdout())
	assert.Equal("this is stdout\n", buf.String())
}

func TestOsExecCommand_StartAndWaitWithStderr(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	stdoutBuf := &strings.Builder{}
	stderrBuf := &strings.Builder{}

	command.SetStdout(stdoutBuf)
	command.SetStderr(stderrBuf)

	err := command.Start()
	assert.NoError(err)

	assert.NoError(command.Wait())

	assert.Same(stdoutBuf, command.GetStdout())
	assert.Equal("this is stdout\n", stdoutBuf.String())
	assert.Same(stderrBuf, command.GetStderr())
	assert.Equal("and this is stderr\n", stderrBuf.String())
}

func TestOsExecCommand_CombinedOutput(t *testing.T) {
	assert := require.New(t)

	command := getCommand()

	output, err := command.CombinedOutput()
	assert.NoError(err)
	assert.Equal("this is stdout\nand this is stderr\n", string(output))
}

func TestOsExecCommand_Output(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", "echo 'this is' stdout ; >&2 echo 'much error' ; exit 123")

	output, err := command.Output()
	assert.EqualError(err, "exit status 123")
	assert.Equal("this is stdout\n", string(output))

	var exitErr cmd.ExitError
	assert.True(errors.As(err, &exitErr))
	assert.Equal("much error\n", string(exitErr.Stderr()))
}

func TestOsExecCommand_Pipes(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", "cat ; >&2 echo 'and stderr'")

	stdin, err := command.StdinPipe()
	assert.NoError(err)

	wg := sync.WaitGroup{}
	wg.Add(2)

	stdout, err := command.StdoutPipe()
	assert.NoError(err)

	stderr, err := command.StderrPipe()
	assert.NoError(err)

	assert.NoError(command.Start())

	stdinContents := "this is from stdin"
	n, err := io.WriteString(stdin, stdinContents)
	assert.Equal(len(stdinContents), n)
	assert.NoError(err)
	assert.NoError(stdin.Close())

	stdoutContent, err := io.ReadAll(stdout)
	assert.NoError(err)

	stderrContent, err := io.ReadAll(stderr)
	assert.NoError(err)

	assert.Equal(stdinContents, string(stdoutContent))
	assert.Equal("and stderr\n", string(stderrContent))

	assert.NoError(command.Wait())
}

func TestOsExecCommand_OtherMethods(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("something", "arguments", "here")

	assert.Equal("something", command.GetPath())
	command.SetPath("changed")
	assert.Equal("changed", command.GetPath())

	assert.Equal([]string{"something", "arguments", "here"}, command.GetArgs())
	command.AppendArgs("another")
	assert.Equal([]string{"something", "arguments", "here", "another"}, command.GetArgs())
	command.SetArgs([]string{"now", "final"})
	assert.Equal([]string{"now", "final"}, command.GetArgs())

	assert.Nil(command.GetEnv())
	command.AppendEnv("LOREM=ipsum", "DOLOR=sit amet")
	assert.Equal([]string{"LOREM=ipsum", "DOLOR=sit amet"}, command.GetEnv())
	command.SetEnv([]string{"WHATEVER=wow"})
	assert.Equal([]string{"WHATEVER=wow"}, command.GetEnv())

	assert.Equal("", command.GetDir())
	command.SetDir("/opt")
	assert.Equal("/opt", command.GetDir())

	reader := strings.NewReader("stdin")

	assert.Nil(command.GetStdin())
	command.SetStdin(reader)
	assert.Same(reader, command.GetStdin())

	file1, err := os.CreateTemp("", "cmd-test")
	assert.NoError(err)

	defer file1.Close()
	defer os.Remove(file1.Name())

	file2, err := os.CreateTemp("", "cmd-test")
	assert.NoError(err)

	defer file2.Close()
	defer os.Remove(file2.Name())

	assert.Nil(command.GetExtraFiles())
	command.AppendExtraFiles(file1)
	assert.Len(command.GetExtraFiles(), 1)
	assert.Same(file1, command.GetExtraFiles()[0])
	command.SetExtraFiles([]*os.File{file2})
	assert.Len(command.GetExtraFiles(), 1)
	assert.Same(file2, command.GetExtraFiles()[0])

	sysProcAttr := &syscall.SysProcAttr{Chroot: "/opt"}

	assert.Nil(command.GetSysProcAttr())
	command.SetSysProcAttr(sysProcAttr)
	assert.Same(sysProcAttr, command.GetSysProcAttr())

	assert.Equal("now final", command.String())
}

func TestOsExecCommand_NotExitError(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("something that doesnt exist")

	err := command.Run()

	var exitErr cmd.ExitError
	assert.False(errors.As(err, &exitErr))
	assert.IsType(&exec.Error{}, err)
}

func TestOsExecCommand_PanicAfterChangingPropertiesAfterStart(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "5")

	assert.NoError(command.Start())

	defer func() {
		command.GetProcess().Kill()
		command.GetProcess().Release()
	}()

	assert.Panics(func() { command.SetPath("whatever") })
	assert.Panics(func() { command.SetArgs([]string{"whatever"}) })
	assert.Panics(func() { command.AppendArgs("whatever") })
	assert.Panics(func() { command.SetEnv([]string{"whatever=lol"}) })
	assert.Panics(func() { command.AppendEnv("whatever") })
	assert.Panics(func() { command.SetStdin(nil) })
	assert.Panics(func() { command.SetStdout(nil) })
	assert.Panics(func() { command.SetStderr(nil) })
	assert.Panics(func() { command.SetExtraFiles(nil) })
	assert.Panics(func() { command.AppendExtraFiles(nil) })
	assert.Panics(func() { command.SetSysProcAttr(nil) })
	assert.Panics(func() { command.SetCancel(nil) })
	assert.Panics(func() { command.SetWaitDelay(time.Second) })
}

func TestOsExecCommand_Cancel(t *testing.T) {
	assert := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := cmd.NewOsExecFactory().CommandContext(ctx, "sleep", "10")

	assert.NotNil(command.GetCancel()) // CommandContext's default: Kill

	var cancelCalled atomic.Bool

	command.SetCancel(func() error {
		cancelCalled.Store(true)

		return command.GetProcess().Signal(os.Interrupt)
	})

	assert.NotNil(command.GetCancel())
	assert.NoError(command.Start())

	cancel()

	assert.EqualError(command.Wait(), "signal: interrupt")
	assert.True(cancelCalled.Load())
}

func TestOsExecCommand_GracefulTermination(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "ready")

	for name, test := range map[string]struct {
		script string
		err    string
	}{
		"exits on SIGTERM":   {script: "touch " + ready + "; exec sleep 10", err: "signal: terminated"},
		"killed after delay": {script: `trap "" TERM; touch ` + ready + "; exec sleep 10", err: "signal: killed"},
	} {
		t.Run(name, func(t *testing.T) {
			assert := require.New(t)

			_ = os.Remove(ready)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			command := cmd.NewOsExecFactory().CommandContext(ctx, "sh", "-c", test.script)
			command.SetCancel(func() error { return command.GetProcess().Signal(syscall.SIGTERM) })
			command.SetWaitDelay(20 * time.Millisecond)

			go func() {
				// Cancel once the script is ready (the trap, if any, is installed).
				for ctx.Err() == nil {
					if _, err := os.Stat(ready); err == nil {
						cancel()

						return
					}

					time.Sleep(time.Millisecond)
				}
			}()

			_, err := command.Output()
			assert.EqualError(err, test.err)
		})
	}
}

func TestOsExecCommand_WaitDelay(t *testing.T) {
	assert := require.New(t)

	// The background sleep keeps stdout open after sh exits; WaitDelay makes Wait stop waiting for it.
	command := cmd.NewOsExecFactory().Command("sh", "-c", "sleep 0.5 &")
	command.SetStdout(&strings.Builder{})

	assert.Equal(time.Duration(0), command.GetWaitDelay())
	command.SetWaitDelay(10 * time.Millisecond)
	assert.Equal(10*time.Millisecond, command.GetWaitDelay())

	err := command.Run()
	assert.ErrorIs(err, exec.ErrWaitDelay)
	assert.True(command.GetProcessState().Success())
}

func TestOsExecCommand_GetErr(t *testing.T) {
	assert := require.New(t)

	factory := cmd.NewOsExecFactory()

	assert.NoError(factory.Command("sh").GetErr())

	command := factory.Command("something-that-doesnt-exist")
	assert.ErrorIs(command.GetErr(), exec.ErrNotFound)
	assert.Equal(command.GetErr(), command.Start())
}

func TestOsExecCommand_Environ(t *testing.T) {
	assert := require.New(t)

	t.Setenv("GO_CMD_TEST", "inherited")

	command := cmd.NewOsExecFactory().Command("sh")
	assert.Contains(command.Environ(), "GO_CMD_TEST=inherited")

	command.SetEnv([]string{"ENVIRON=explicit"})
	assert.Equal([]string{"ENVIRON=explicit"}, command.Environ())
}

func TestOsExecCommand_GetProcessStateIsNilUntilFinished(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sleep", "0.01")

	assert.True(command.GetProcessState() == nil) // a real nil interface, not a nil *os.ProcessState

	assert.NoError(command.Start())
	assert.True(command.GetProcessState() == nil)

	assert.NoError(command.Wait())

	state := command.GetProcessState()
	assert.IsType(&cmd.OsProcessState{}, state)
	assert.True(state.Success())
}

func TestOsExecCommand_WaitRepeatedly(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("sh", "-c", "exit 4")

	assert.EqualError(command.Wait(), "exec: not started")

	err := command.Run()
	assert.EqualError(err, "exit status 4")

	// Same as exec.Cmd: the second Wait reports the misuse.
	assert.EqualError(command.Wait(), "exec: Wait was already called")
}

// results of running the same command through exec.Cmd (raw) and through the wrapper.
type results struct {
	rawOut, out []byte
	rawErr, err error
}

// TestOsExecCommand_SameResultsAsExec runs the same commands through exec.Cmd and through the wrapper and
// compares the results.
func TestOsExecCommand_SameResultsAsExec(t *testing.T) {
	scripts := map[string]string{
		"success":            "echo out; echo err >&2",
		"exit code":          "echo out; echo err >&2; exit 3",
		"long stderr":        `head -c 100000 /dev/zero | tr '\0' x >&2; exit 1`,
		"killed by a signal": "kill -KILL $$",
	}

	for name, script := range scripts {
		t.Run(name, func(t *testing.T) {
			for method, run := range map[string]func(*exec.Cmd, cmd.Command) results{
				"Output": func(raw *exec.Cmd, wrapped cmd.Command) results {
					var r results

					r.rawOut, r.rawErr = raw.Output()
					r.out, r.err = wrapped.Output()

					return r
				},
				"CombinedOutput": func(raw *exec.Cmd, wrapped cmd.Command) results {
					var r results

					r.rawOut, r.rawErr = raw.CombinedOutput()
					r.out, r.err = wrapped.CombinedOutput()

					return r
				},
				"Run": func(raw *exec.Cmd, wrapped cmd.Command) results {
					return results{rawErr: raw.Run(), err: wrapped.Run()}
				},
			} {
				t.Run(method, func(t *testing.T) {
					assert := require.New(t)

					raw := exec.Command("sh", "-c", script)
					wrapped := cmd.NewOsExecFactory().Command("sh", "-c", script)

					r := run(raw, wrapped)
					rawErr, err, rawOut, out := r.rawErr, r.err, r.rawOut, r.out

					assert.Equal(rawOut, out)
					assert.Equal(raw.ProcessState.String(), wrapped.GetProcessState().String())

					if rawErr == nil {
						assert.NoError(err)

						return
					}

					assert.EqualError(err, rawErr.Error())

					var rawExitErr *exec.ExitError
					assert.True(errors.As(rawErr, &rawExitErr))

					var exitErr cmd.ExitError
					assert.True(errors.As(err, &exitErr))
					assert.Equal(rawExitErr.Stderr, exitErr.Stderr())
					assert.Equal(rawExitErr.ExitCode(), exitErr.ExitCode())
				})
			}
		})
	}
}

// TestOsExecCommand_WrappedExitErrorIsUnchanged checks that only a bare *exec.ExitError is converted: an error
// that merely wraps one (here returned by Cancel) is returned exactly like exec.Cmd returns it.
func TestOsExecCommand_WrappedExitErrorIsUnchanged(t *testing.T) {
	assert := require.New(t)

	otherErr := exec.Command("sh", "-c", "exit 7").Run()

	var otherExitErr *exec.ExitError
	assert.True(errors.As(otherErr, &otherExitErr))

	run := func(command interface {
		Start() error
		Wait() error
	}, cancel context.CancelFunc,
	) error {
		assert.NoError(command.Start())
		cancel()

		return command.Wait()
	}

	// The process ignores the cancellation and exits 0; exec.Cmd then reports the error returned by Cancel.
	script := `trap "" TERM INT; sleep 0.05`

	rawCtx, rawCancel := context.WithCancel(context.Background())
	defer rawCancel()

	raw := exec.CommandContext(rawCtx, "sh", "-c", script)
	raw.Cancel = func() error { return otherErr }
	rawErr := run(raw, rawCancel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wrapped := cmd.NewOsExecFactory().CommandContext(ctx, "sh", "-c", script)
	wrapped.SetCancel(func() error { return otherErr })
	err := run(wrapped, cancel)

	assert.EqualError(rawErr, "exec: canceling Cmd: exit status 7")
	assert.EqualError(err, rawErr.Error())
	assert.IsType(rawErr, err)
	assert.ErrorIs(err, otherErr)
	assert.True(wrapped.GetProcessState().Success())
}

// TestOsExecCommand_GetProcessFromCancel calls GetProcess from a Cancel function (which runs on its own
// goroutine) while the test calls it too; -race reports any unsynchronized access.
func TestOsExecCommand_GetProcessFromCancel(t *testing.T) {
	assert := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := cmd.NewOsExecFactory().CommandContext(ctx, "sleep", "10")
	command.SetCancel(func() error { return command.GetProcess().Kill() })

	assert.NoError(command.Start())

	cancel()

	for range 100 {
		assert.NotNil(command.GetProcess())
	}

	assert.EqualError(command.Wait(), "signal: killed")
}

func TestOsExecCommand_SetErr(t *testing.T) {
	assert := require.New(t)

	command := cmd.NewOsExecFactory().Command("something-that-doesnt-exist")
	assert.ErrorIs(command.GetErr(), exec.ErrNotFound)

	// Start fails with Err (if set), so replacing or clearing it (the os/exec pattern for accepting ErrDot)
	// changes what Start does.
	errCustom := errors.New("custom")

	command.SetErr(errCustom)
	assert.Same(errCustom, command.GetErr())
	assert.Same(errCustom, command.Start())

	command = cmd.NewOsExecFactory().Command("something-that-doesnt-exist")
	command.SetErr(nil)
	assert.NoError(command.GetErr())

	err := command.Start()
	assert.Error(err)
	assert.NotErrorIs(err, exec.ErrNotFound)
}

func TestOsExecCommand_Cmd(t *testing.T) {
	assert := require.New(t)

	raw := exec.Command("true")
	command := cmd.NewOsExecCommand(raw)

	assert.Same(raw, command.Cmd())
}

func TestOsExecCommand_OutputsWithStdioAlreadySet(t *testing.T) {
	assert := require.New(t)

	command := getCommand()
	command.SetStdout(&strings.Builder{})

	_, err := command.Output()
	assert.EqualError(err, "exec: Stdout already set")

	_, err = command.CombinedOutput()
	assert.EqualError(err, "exec: Stdout already set")

	command = getCommand()
	command.SetStderr(&strings.Builder{})

	_, err = command.CombinedOutput()
	assert.EqualError(err, "exec: Stderr already set")
}

func getCommand() *cmd.OsExecCommand {
	command := "echo 'this is' stdout ; >&2 echo 'and this is' stderr"

	return cmd.NewOsExecFactory().Command("sh", "-c", command).(*cmd.OsExecCommand)
}
