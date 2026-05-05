package goivy

import "os"

// WithWorkingDir temporarily changes the working directory, restoring it on return.
// Corresponds to Python's WorkingDir context manager.
func WithWorkingDir(dir string, fn func()) error {
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer os.Chdir(oldDir)
	fn()
	return nil
}
