package ivyutils

import "os"

// Filename is the current source file being processed.
// Corresponds to Python's module-level: filename = None
var Filename string

// WithSourceFile temporarily sets the global Filename, restoring the old value on return.
// Corresponds to Python's SourceFile context manager.
func WithSourceFile(fname string, fn func()) {
	oldFilename := Filename
	Filename = fname
	defer func() { Filename = oldFilename }()
	fn()
}

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
