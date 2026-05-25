package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- M6: class target tests ------------------------------------------

// TestClassTarget_NoMainNoRepl confirms the class target produces a
// library-shaped package: no main.go, no repl.go, but state.go +
// actions.go + runtime.go are present and exported.
func TestClassTarget_NoMainNoRepl(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, never := range []string{"main.go", "repl.go"} {
		if _, has := out.Files[never]; has {
			t.Errorf("class target should not emit %s", never)
		}
	}
	for _, want := range []string{"state.go", "init.go", "runtime.go", "actions.go"} {
		if _, has := out.Files[want]; !has {
			t.Errorf("class target missing %s; keys=%v", want, keysOf(out.Files))
		}
	}
}

// TestClassTarget_ExportsStateAPI confirms NewState and action
// methods are exported (capitalised) so a downstream Go consumer can
// call them.
func TestClassTarget_ExportsStateAPI(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	state := out.Files["state.go"]
	if !strings.Contains(state, "type State struct") {
		t.Errorf("state.go missing exported State struct:\n%s", state)
	}
	if !strings.Contains(state, "func NewState() *State") {
		t.Errorf("state.go missing exported NewState:\n%s", state)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "func (s *State) SetFlag()") {
		t.Errorf("actions.go missing exported SetFlag method:\n%s", actions)
	}
	init := out.Files["init.go"]
	if !strings.Contains(init, "func (s *State) Init()") {
		t.Errorf("init.go missing exported Init:\n%s", init)
	}
}

// TestSmoke_BuildEmittedClass builds the emitted library package
// (no main, no binary) via `go build`. Gated by SLOW_GO_TEST.
func TestSmoke_BuildEmittedClass(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"),
		[]byte("module ivygo_class_smoke\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed:\n%s", string(output))
	}
}

// TestClassTarget_CardinalityTypes ensures a class output with mixed
// type declarations (enum, range, uninterpreted) still go-builds.
func TestSmoke_BuildEmittedClass_MixedTypes(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type color = {red, green, blue}
type idx = {0..3}
relation pick(I: idx, C: color)
action set_red = {
	pick(0, red) := true
}
`)
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"),
		[]byte("module ivygo_class_mixed\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed:\n%s", string(output))
	}
}
