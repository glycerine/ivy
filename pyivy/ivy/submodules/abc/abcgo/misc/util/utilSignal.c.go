package util

import (
	"os"
	"os/exec"
	"runtime"
)

// Util_SignalSystem preserves system()'s shell-command behavior and Unix-style
// wait status: normal exit codes occupy the upper eight bits.
func Util_SignalSystem(command string) int32 {
	shell, args := "sh", []string{"-c", command}
	if runtime.GOOS == "windows" {
		shell, args = "cmd", []string{"/C", command}
	}
	cmd := exec.Command(shell, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		if exit.ExitCode() >= 0 {
			return int32(exit.ExitCode() << 8)
		}
		return 1
	}
	return -1
}
func Util_SignalTmpFile(prefix, suffix string, outName *string) (*os.File, error) {
	return tmpFile(prefix, suffix, outName)
}
func Util_SignalTmpFileRemove(name string, leave int32) {
	if leave == 0 {
		_ = os.Remove(name)
	}
}
