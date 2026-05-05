package goivy

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func pythonIvyCommandForTest(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()

	repoRoot := testRepoRoot(t)
	pyivyRoot := filepath.Join(repoRoot, "pyivy", "ivy")
	python := filepath.Join(repoRoot, "pyivy", "goivy-venv", "bin", "python3")
	if _, err := os.Stat(python); err != nil {
		if _, lookErr := exec.LookPath("python3"); lookErr != nil {
			t.Skip("python3 not available")
		}
		python = "python3"
	}

	cmd := exec.Command(python, args...)
	cmd.Dir = pyivyRoot
	pythonPath := pyivyRoot
	if existing := os.Getenv("PYTHONPATH"); existing != "" {
		pythonPath += string(os.PathListSeparator) + existing
	}
	cmd.Env = append(os.Environ(),
		"IVY_HOME="+pyivyRoot,
		"PYTHONPATH="+pythonPath,
	)
	return cmd
}
