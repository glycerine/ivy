package abc

// Process utilities from src/misc/util/utilSignal.c and the temporary-file
// implementation shared with src/misc/util/utilFile.c.

import (
	"os"
	"os/exec"
	"path/filepath"
)

// UtilSignalSystem executes command with the pinned non-Windows system shell
// and returns the wait-status encoding used by C system: exit code << 8. A
// shell start failure or signal-only status returns -1.
func UtilSignalSystem(command string) int {
	err := exec.Command("/bin/sh", "-c", command).Run()
	if err == nil {
		return 0
	}
	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() < 0 {
		return -1
	}
	return exitError.ExitCode() << 8
}

// CreateSignalTempFile atomically creates a mode-0600 temporary file whose
// path begins with prefix and ends with suffix. The caller owns both file and
// path. It replaces native's unlink-and-reopen suffix kludge with one atomic
// creation.
func CreateSignalTempFile(prefix, suffix string) (*os.File, string, error) {
	directory, base := filepath.Split(prefix)
	if directory == "" {
		directory = "."
	}
	file, err := os.CreateTemp(directory, base+"*"+suffix)
	if err != nil {
		return nil, "", err
	}
	return file, file.Name(), nil
}

// RemoveSignalTempFile removes name unless leave is true. Unlike the native
// void wrapper, removal failures are returned to the caller.
func RemoveSignalTempFile(name string, leave bool) error {
	if leave {
		return nil
	}
	return os.Remove(name)
}
