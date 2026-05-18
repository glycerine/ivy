package goivy

// PythonOracle manages a long-running Python subprocess that parses
// Ivy expressions using the real Python Ivy parser and returns
// canonical AST shape strings for cross-validation.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// PythonOracle is a long-running Python subprocess for cross-validation.
// It communicates via a line protocol over stdin/stdout pipes.
type PythonOracle struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	mu     sync.Mutex // serializes access to the subprocess
}

var (
	oracleOnce     sync.Once
	oracleInstance *PythonOracle
	oracleErr      error
)

// pythonScriptPath returns the path to ivy_expr_shape.py.
func pythonScriptPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "pytesthelper", "ivy_expr_shape.py")
}

// pyivyDir returns the path to the Python Ivy source tree.
func pyivyDir() string {
	if root := os.Getenv("PYIVY_ROOT"); root != "" {
		return root
	}
	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, "ivy", "pyivy", "ivy"),
		filepath.Join(home, "pyivy", "ivy"),
	}
	for _, candidate := range candidates {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate
		}
	}
	return candidates[0]
}

func pyivyPython() string {
	if py := os.Getenv("PYIVY_PYTHON"); py != "" {
		return py
	}
	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, "ivy", "pyivy", "goivy-venv", "bin", "python3"),
		filepath.Join(home, "pyivy", "goivy-venv", "bin", "python3"),
	}
	for _, candidate := range candidates {
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	return "python3"
}

// newPythonOracle starts the Python oracle subprocess.
func newPythonOracle() (*PythonOracle, error) {
	script := pythonScriptPath()
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("python oracle script not found: %s", script)
	}

	ivyDir := pyivyDir()
	if _, err := os.Stat(ivyDir); err != nil {
		return nil, fmt.Errorf("pyivy directory not found: %s", ivyDir)
	}

	// Use -O to suppress __debug__/xtracer output
	cmd := exec.Command(pyivyPython(), "-O", script)
	cmd.Dir = ivyDir
	pythonPath := ivyDir
	if existing := os.Getenv("PYTHONPATH"); existing != "" {
		pythonPath += string(os.PathListSeparator) + existing
	}
	cmd.Env = append(os.Environ(), "PYTHONPATH="+pythonPath)
	cmd.Stderr = os.Stderr // let Python errors through for debugging

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("starting python oracle: %w", err)
	}

	reader := bufio.NewReader(stdout)

	// Wait for READY signal
	line, err := reader.ReadString('\n')
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("reading READY from python oracle: %w", err)
	}
	line = strings.TrimSpace(line)
	if line != "READY" {
		cmd.Process.Kill()
		return nil, fmt.Errorf("expected READY, got: %q", line)
	}

	return &PythonOracle{
		cmd:    cmd,
		stdin:  stdin,
		reader: reader,
	}, nil
}

// getOracle returns the singleton PythonOracle, starting it if needed.
func getOracle(t *testing.T) *PythonOracle {
	t.Helper()
	oracleOnce.Do(func() {
		oracleInstance, oracleErr = newPythonOracle()
	})
	if oracleErr != nil {
		t.Skipf("python oracle unavailable: %v", oracleErr)
	}
	return oracleInstance
}

// send sends a command and reads the response line.
func (o *PythonOracle) send(cmd string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	_, err := fmt.Fprintf(o.stdin, "%s\n", cmd)
	if err != nil {
		return "", fmt.Errorf("writing to python oracle: %w", err)
	}

	line, err := o.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading from python oracle: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// ParseExpr parses an expression with the Python Ivy parser.
// Returns (shape, nil) on success or ("", error) on failure.
func (o *PythonOracle) ParseExpr(input string) (string, error) {
	resp, err := o.send("EXPR " + input)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(resp, "OK ") {
		return resp[3:], nil
	}
	if strings.HasPrefix(resp, "ERR ") {
		return "", fmt.Errorf("python parse error: %s", resp[4:])
	}
	return "", fmt.Errorf("unexpected response: %q", resp)
}

// ParseAction parses an action body with the Python Ivy parser.
// Returns (shape, nil) on success or ("", error) on failure.
func (o *PythonOracle) ParseAction(input string) (string, error) {
	resp, err := o.send("ACTION " + input)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(resp, "OK ") {
		return resp[3:], nil
	}
	if strings.HasPrefix(resp, "ERR ") {
		return "", fmt.Errorf("python parse error: %s", resp[4:])
	}
	return "", fmt.Errorf("unexpected response: %q", resp)
}

// ParseExprBatch sends multiple expressions in batch mode.
// Returns a slice of (shape, error) pairs, one per input.
//
// Writing and reading happen concurrently to avoid deadlock:
// the OS pipe buffer (~64KB) can fill if we try to write all
// inputs before reading any responses.
func (o *PythonOracle) ParseExprBatch(inputs []string) []batchResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	results := make([]batchResult, len(inputs))

	// Write all inputs in a goroutine to avoid pipe buffer deadlock
	go func() {
		fmt.Fprintf(o.stdin, "BATCH\n")
		for _, input := range inputs {
			fmt.Fprintf(o.stdin, "EXPR %s\n", input)
		}
		fmt.Fprintf(o.stdin, "END\n")
	}()

	// Read all responses in the main goroutine
	for i := range inputs {
		line, err := o.reader.ReadString('\n')
		if err != nil {
			results[i] = batchResult{err: fmt.Errorf("reading batch response %d: %w", i, err)}
			// Fill remaining with the same error
			for j := i + 1; j < len(inputs); j++ {
				results[j] = batchResult{err: fmt.Errorf("skipped after read error at %d", i)}
			}
			break
		}
		resp := strings.TrimSpace(line)
		if strings.HasPrefix(resp, "OK ") {
			results[i] = batchResult{shape: resp[3:]}
		} else if strings.HasPrefix(resp, "ERR ") {
			results[i] = batchResult{err: fmt.Errorf("python parse error: %s", resp[4:])}
		} else {
			results[i] = batchResult{err: fmt.Errorf("unexpected response: %q", resp)}
		}
	}
	return results
}

type batchResult struct {
	shape string
	err   error
}

// Close shuts down the Python subprocess.
func (o *PythonOracle) Close() {
	if o.stdin != nil {
		o.stdin.Close()
	}
	if o.cmd != nil && o.cmd.Process != nil {
		o.cmd.Wait()
	}
}
