package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- M7: REPL emission tests ----------------------------------------

func TestEmitRepl_ProducesReplFile(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, has := out.Files["repl.go"]; !has {
		t.Fatalf("repl target should emit repl.go; keys=%v", keysOf(out.Files))
	}
}

func TestEmitRepl_DispatchTableHasAction(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["repl.go"]
	if !strings.Contains(text, `case "set_flag":`) {
		t.Errorf("repl dispatch should include case for set_flag, got:\n%s", text)
	}
	if !strings.Contains(text, "state.SetFlag(") {
		t.Errorf("repl dispatch should call state.SetFlag, got:\n%s", text)
	}
}

func TestEmitRepl_BoolArgParser(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag(b: bool) = {
	flag := b
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["repl.go"]
	if !strings.Contains(text, "parseArg_bool") {
		t.Errorf("repl should emit bool arg parser, got:\n%s", text)
	}
}

func TestEmitRepl_EnumArgParserMatchesByName(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green, blue}
function pick : color
action set_pick(c: color) = {
	pick := c
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["repl.go"]
	if !strings.Contains(text, "parseArg_Color") {
		t.Errorf("repl missing Color arg parser, got:\n%s", text)
	}
	// gofmt splits `case "red": return Red, nil` across lines;
	// look for the `case` and the `Red` return separately.
	if !strings.Contains(text, `case "red":`) {
		t.Errorf("repl should map 'red' token to a case:\n%s", text)
	}
	if !strings.Contains(text, "return Red, nil") {
		t.Errorf("repl should return Red enum constant:\n%s", text)
	}
}

func TestEmitMain_ReplTargetCallsRunRepl(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["main.go"]
	if !strings.Contains(text, "state := NewState()") {
		t.Errorf("main should construct state, got:\n%s", text)
	}
	if !strings.Contains(text, "state.Init()") {
		t.Errorf("main should call Init, got:\n%s", text)
	}
	if !strings.Contains(text, "runRepl(state, os.Stdin, os.Stdout)") {
		t.Errorf("main should drive runRepl, got:\n%s", text)
	}
}

func TestEmitMain_ImplTargetSkipsRunRepl(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["main.go"]
	if strings.Contains(text, "runRepl") {
		t.Errorf("impl main should not call runRepl, got:\n%s", text)
	}
}

func TestEmitRepl_FileIsGofmtClean(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green, blue}
function pick : color
action set_pick(c: color) = {
	pick := c
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

// --- M7: Tier 2 smoke — repl target go-builds ----------------------

func TestSmoke_BuildEmittedRepl_FlagAction(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
action unset = {
	flag := false
}
`)
	out, err := Generate(mod, Config{Target: "repl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"),
		[]byte("module ivygo_repl_smoke\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(pkgDir, "repl_bin"), "./...")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed:\n%s", string(output))
	}
}
