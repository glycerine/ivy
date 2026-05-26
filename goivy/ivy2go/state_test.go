package ivy2go

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// --- M5: state emission tests ---------------------------------------

func TestEmitState_RelationEmitsBoolField(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["state.go"]
	if text == "" {
		t.Fatalf("state.go missing:\n%v", keysOf(out.Files))
	}
	requireHasLineWithAllTerms(t, text, "type", "State", "struct")
	requireHasLineWithAllTerms(t, text, "Flag", "bool")
	requireHasLineWithAllTerms(t, text, "func NewState() *State")
}

func TestEmitState_FunctionRelationEmitsArrayOrMap(t *testing.T) {
	mod := compileIvySource(t, `
type node = {0..3}
relation link(N1: node, N2: node)
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["state.go"]
	// Two-arg over 4-elem domains → array storage `[4][4]bool`.
	// gofmt aligns adjacent struct fields, so accept any whitespace
	// between `Link` and `[4][4]bool` (single space when there are
	// no co-aligned fields, multiple spaces when there are).
	if !regexp.MustCompile(`Link\s+\[4\]\[4\]bool`).MatchString(text) {
		t.Errorf("expected `Link [4][4]bool` field, got:\n%s", text)
	}
}

func TestEmitState_LargeDomainUsesMap(t *testing.T) {
	mod := compileIvySource(t, `
type node = {0..2048}
function size(N: node) : node
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["state.go"]
	// Domain card 2049 is past largeThresh (1024) → map storage.
	if !strings.Contains(text, "map[Node]Node") {
		t.Errorf("expected map[Node]Node, got:\n%s", text)
	}
}

func TestEmitNewState_AllocatesMaps(t *testing.T) {
	mod := compileIvySource(t, `
type node = {0..2048}
function size(N: node) : node
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["state.go"]
	if !strings.Contains(text, "s.Size = map[Node]Node{}") {
		t.Errorf("NewState should allocate map field, got:\n%s", text)
	}
}

// --- M5: Init emission tests ----------------------------------------

func TestEmitInit_NondetScalar(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["init.go"]
	requireHasLineWithAllTerms(t, text, "func (s *State) Init()")
	// Scalar boolean → ivyChoose(2) == 1.
	if !strings.Contains(text, "s.Flag = ivyChoose(2) == 1") {
		t.Errorf("expected nondet flag init, got:\n%s", text)
	}
}

func TestEmitInit_TestTargetEmitsNondetWhenNoAxioms(t *testing.T) {
	// Mirrors ivy2cpp/initial_state.go emitDefaultInitialState: when
	// the module has no InitCond formulas to solve, the test/gen
	// targets fall back to nondet just like impl/repl — they don't
	// silently leave fields at the Go zero value.
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["init.go"]
	if strings.Contains(text, "TODO(M9)") {
		t.Errorf("test target with no InitCond should not emit M9 TODO, got:\n%s", text)
	}
	if !strings.Contains(text, "s.Flag = ivyChoose(2) == 1") {
		t.Errorf("expected nondet flag init under test target, got:\n%s", text)
	}
}


// --- M5: build plan tests -------------------------------------------

func TestBuildPlanFor_ImplProducesBinary(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	plan, err := BuildPlanFor(out, t.TempDir())
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if plan.CompileOnly {
		t.Error("impl target should not be CompileOnly")
	}
	if plan.OutputPath == "" {
		t.Error("impl plan should set OutputPath")
	}
	if len(plan.Args) < 2 || plan.Args[0] != "build" {
		t.Errorf("plan.Args = %v, want [build, -o, ...]", plan.Args)
	}
}

func TestBuildPlanFor_ClassCompileOnly(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	plan, err := BuildPlanFor(out, t.TempDir())
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if !plan.CompileOnly {
		t.Error("class target should be CompileOnly")
	}
}

// --- M5: Tier 2 build smoke test (SLOW_GO_TEST gated) ---------------

func TestSmoke_BuildEmittedPackage_EmptyModule(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}

func TestSmoke_BuildEmittedPackage_RelationFlag(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}

// --- Always-on Tier 3 across M5: every emitted file remains valid ---

func TestEmitState_AllFilesGofmtClean(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green, blue}
type node = {0..3}
relation flag
relation link(N1: node, N2: node)
function pick : color
action set_pick = {
	pick := red
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}
