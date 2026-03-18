package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PyBackend communicates with a Python Ivy sidecar HTTP server.
type PyBackend struct {
	cmd     *exec.Cmd
	baseURL string
	client  *http.Client
	mu      sync.RWMutex
}

// pyIvyRoot returns the root of the Python Ivy source tree.
// It checks PYIVY_ROOT env var, then falls back to ~/pyivy/ivy.
func pyIvyRoot() string {
	if root := os.Getenv("PYIVY_ROOT"); root != "" {
		return root
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("pyivy", "ivy")
	}
	return filepath.Join(home, "pyivy", "ivy")
}

// pyIvyPython returns the Python interpreter to use for the sidecar.
// Priority: PYIVY_PYTHON env var > ~/pyivy/venv/bin/python3 > python3.
func pyIvyPython() string {
	if py := os.Getenv("PYIVY_PYTHON"); py != "" {
		return py
	}
	home, err := os.UserHomeDir()
	if err == nil {
		venvPy := filepath.Join(home, "pyivy", "venv", "bin", "python3")
		if _, err := os.Stat(venvPy); err == nil {
			return venvPy
		}
	}
	return "python3"
}

// NewPyBackend starts the Python sidecar and returns a PyBackend.
// It finds a free port, launches the sidecar, and waits for it to be ready.
//
// To make sure your venv has ivy from ~/pyivy installed:
// source ~/pyivy/venv/bin/activate
// cd ~/pyivy/ivy
// pip install -e .
func NewPyBackend() (*PyBackend, error) {
	// Find a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	root := pyIvyRoot()

	// The Python Ivy source tree includes a symlink ivy/z3 -> build/.../ivy/z3
	// which contains the bundled Z3 4.7.1 fork (libz3.dylib + Python bindings).
	// DYLD_LIBRARY_PATH must point to ivy/z3 so ctypes finds libz3.dylib.
	z3Dir := filepath.Join(root, "ivy", "z3")

	// Verify the z3 directory exists (may be a symlink to the build output).
	if _, err := os.Stat(z3Dir); err != nil {
		return nil, fmt.Errorf("python ivy z3 not found at %s: %w\n"+
			"Ensure the Z3 build exists. Try:\n"+
			"  cd %s && python3 build_submodules.py\n"+
			"  ln -s build/lib.*/ivy/z3 ivy/z3", z3Dir, err, root)
	}

	// Start the Python sidecar.
	// Use the venv Python if available so we get the correct ivy version.
	pythonBin := pyIvyPython()

	// now use locally vendored version for more independence of pyivy
	cmd := exec.Command(pythonBin, "../pytesthelper/sidecar.py", "--port", fmt.Sprint(port))

	// Build environment with path to bundled Z3 library.
	env := os.Environ()
	env = appendEnvPath(env, "DYLD_LIBRARY_PATH", z3Dir)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("pybackend: starting sidecar on port %d (python=%s, root=%s)", port, pythonBin, root)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start python sidecar: %w", err)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	pb := &PyBackend{
		cmd:     cmd,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}

	// Wait for sidecar to become ready.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := pb.client.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return pb, nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	cmd.Process.Kill()
	return nil, fmt.Errorf("python sidecar did not become ready within 30s")
}

// appendEnvPath prepends dir to an environment variable in an env slice.
func appendEnvPath(env []string, key, dir string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + dir + ":" + e[len(prefix):]
			return env
		}
	}
	return append(env, key+"="+dir)
}

func (b *PyBackend) Close() error {
	if b.cmd != nil && b.cmd.Process != nil {
		b.cmd.Process.Kill()
		b.cmd.Wait()
	}
	return nil
}

// post sends a POST request and returns the response body bytes.
func (b *PyBackend) post(path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(data))
	}
	resp, err := b.client.Post(b.baseURL+path, "application/json", bodyReader)
	if err != nil {
		return nil, fmt.Errorf("python backend: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("python backend read: %w", err)
	}
	if resp.StatusCode >= 400 {
		// Extract error message from JSON response if possible.
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("python backend: status %d", resp.StatusCode)
	}
	return respBody, nil
}

// get sends a GET request and returns the response body bytes.
func (b *PyBackend) get(path string) ([]byte, error) {
	resp, err := b.client.Get(b.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("python backend: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("python backend read: %w", err)
	}
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("python backend: status %d", resp.StatusCode)
	}
	return respBody, nil
}

func (b *PyBackend) NewSession() ([]byte, error) {
	return b.post("/session/new", nil)
}

func (b *PyBackend) Load(sessionID, filename string, content []byte) ([]byte, error) {
	return b.post("/session/"+sessionID+"/load", map[string]interface{}{
		"filename": filename,
		"content":  string(content),
	})
}

func (b *PyBackend) LoadPath(sessionID, path string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/load", map[string]string{"path": path})
}

func (b *PyBackend) Action(sessionID, action string, args map[string]interface{}) ([]byte, error) {
	return b.post("/session/"+sessionID+"/action", map[string]interface{}{
		"action": action,
		"args":   args,
	})
}

func (b *PyBackend) GetARG(sessionID string) ([]byte, error) {
	return b.get("/session/" + sessionID + "/arg")
}

func (b *PyBackend) GetConcept(sessionID string) ([]byte, error) {
	return b.get("/session/" + sessionID + "/concept")
}

func (b *PyBackend) ConceptSplit(sessionID, concept, splitBy string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/split", map[string]string{
		"concept":  concept,
		"split_by": splitBy,
	})
}

func (b *PyBackend) ConceptEmpty(sessionID, concept string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/empty", map[string]string{"concept": concept})
}

func (b *PyBackend) ConceptRemove(sessionID, concept string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/remove", map[string]string{"concept": concept})
}

func (b *PyBackend) ConceptUndo(sessionID string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/undo", nil)
}

func (b *PyBackend) ConceptMaterialize(sessionID, concept string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/materialize", map[string]string{"concept": concept})
}

func (b *PyBackend) ConceptReset(sessionID string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/reset", nil)
}

func (b *PyBackend) ConceptDiagram(sessionID string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/diagram", nil)
}

func (b *PyBackend) ConceptProjection(sessionID, name, concept string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/concept/projection", map[string]string{
		"name":    name,
		"concept": concept,
	})
}

func (b *PyBackend) GetToggles(sessionID string) ([]byte, error) {
	return b.get("/session/" + sessionID + "/toggles")
}

func (b *PyBackend) SetToggle(sessionID, edge, displayClass string, value bool) ([]byte, error) {
	return b.post("/session/"+sessionID+"/toggles", map[string]interface{}{
		"edge":          edge,
		"display_class": displayClass,
		"value":         value,
	})
}

func (b *PyBackend) Check(sessionID, mode string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/check", map[string]string{"mode": mode})
}

func (b *PyBackend) GetProof(sessionID string) ([]byte, error) {
	return b.get("/session/" + sessionID + "/proof")
}

func (b *PyBackend) ArgAction(sessionID, node, action string, args map[string]interface{}) ([]byte, error) {
	return b.post("/session/"+sessionID+"/arg/action", map[string]interface{}{
		"node":   node,
		"action": action,
		"args":   args,
	})
}

func (b *PyBackend) ProofAction(sessionID, goal, action string) ([]byte, error) {
	return b.post("/session/"+sessionID+"/proof/action", map[string]string{
		"goal":   goal,
		"action": action,
	})
}

func (b *PyBackend) Save(sessionID string) ([]byte, error) {
	return b.get("/session/" + sessionID + "/save")
}

func (b *PyBackend) Events(sessionID string) (<-chan Event, error) {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		resp, err := b.client.Get(b.baseURL + "/session/" + sessionID + "/events")
		if err != nil {
			return
		}
		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)
		for {
			var evt Event
			if err := dec.Decode(&evt); err != nil {
				return
			}
			ch <- evt
		}
	}()
	return ch, nil
}
