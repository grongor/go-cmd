// Package cmd wraps os/exec behind interfaces so that code which runs external commands can be unit-tested
// without running them.
//
// Production code depends on [Factory] (instead of calling [exec.Command] directly) and gets an [OsExecFactory]
// from [NewOsExecFactory]; tests inject the mockery-generated testify mocks from the
// github.com/grongor/go-cmd/mocks package instead.
//
// Every wrapped method behaves exactly like its os/exec counterpart. The only differences are in the types:
//
//   - exec.Cmd fields are accessed through getters and setters (GetDir/SetDir, ...). The setters panic with
//     [ErrAlreadyStarted] once the process has been started; mutating a started exec.Cmd is a bug anyway.
//   - An *exec.ExitError returned by Run, Wait, Output or CombinedOutput is returned as an [ExitError]
//     (an [*OsExecExitError], which unwraps to the original *exec.ExitError). Any other error is returned
//     unchanged.
//   - *os.Process and *os.ProcessState are returned as [Process] and [ProcessState].
//
// When something isn't covered by the interfaces, [OsExecCommand.Cmd] returns the underlying *exec.Cmd.
package cmd
