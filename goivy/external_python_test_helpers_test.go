package goivy_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func externalTestRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func externalPythonIvyCommandForTest(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()

	repoRoot := externalTestRepoRoot(t)
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
