// Package ivyshell provides shell environment setup for compiled Ivy executables.
// This is a port of Python's ivy_shell.py (12 lines).
//
// On macOS, it prints the DYLD_LIBRARY_PATH needed for Z3 and other
// shared libraries. On other platforms, it prints the equivalent LD_LIBRARY_PATH.
package ivyshell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LibDirs returns the library directories needed for compiled Ivy programs.
func LibDirs() []string {
	var dirs []string
	// Check for Z3 library path
	if z3lib := os.Getenv("Z3_LIB_DIR"); z3lib != "" {
		dirs = append(dirs, z3lib)
	}
	// Check standard locations
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, "lib"))
	}
	dirs = append(dirs, "/usr/local/lib")
	return dirs
}

// ShellSetupCommand returns the shell command to set up the library path.
func ShellSetupCommand() string {
	dirs := LibDirs()
	libPaths := make([]string, 0, len(dirs))
	for _, d := range dirs {
		libDir := filepath.Join(d, "lib")
		if _, err := os.Stat(libDir); err == nil {
			libPaths = append(libPaths, libDir)
		} else if _, err := os.Stat(d); err == nil {
			libPaths = append(libPaths, d)
		}
	}
	path := strings.Join(libPaths, ":")

	var pvar string
	if runtime.GOOS == "darwin" {
		pvar = "DYLD_LIBRARY_PATH"
	} else {
		pvar = "LD_LIBRARY_PATH"
	}

	existing := os.Getenv(pvar)
	if existing != "" {
		path += ":" + existing
	}

	return fmt.Sprintf("export %s=%s", pvar, path)
}

// Main prints the library path setup command.
func Main() {
	fmt.Println(ShellSetupCommand())
}
