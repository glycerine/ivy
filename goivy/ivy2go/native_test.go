package ivy2go

import (
	"strings"
	"testing"
)

// --- M10: native go-block emission tests ----------------------------

func TestEmitNative_GoBlockLandsInNativeFile(t *testing.T) {
	mod := compileIvySource(t, "<<<\ngo\n// hand-written Go helper\nfunc handCrafted() {}\n>>>")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["native.go"]
	if text == "" {
		t.Fatalf("native.go should be emitted, keys=%v", keysOf(out.Files))
	}
	if !strings.Contains(text, "func handCrafted()") {
		t.Errorf("native body should land verbatim:\n%s", text)
	}
}

func TestEmitNative_CppBlockIsSkipped(t *testing.T) {
	// A cpp-tagged native block should NOT appear anywhere in the
	// emitted Go package.
	mod := compileIvySource(t, "<<<\ncpp\n// C++ specific\nvoid x() {}\n>>>")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		if strings.Contains(text, "void x()") {
			t.Errorf("%s should not contain cpp native body:\n%s", name, text)
		}
	}
}

func TestEmitNative_GoHeaderLandsInTypesFile(t *testing.T) {
	mod := compileIvySource(t, "<<<\ngo_header\nconst HelperConst = 42\n>>>")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if text == "" {
		t.Fatalf("types.go should be emitted for go_header block")
	}
	if !strings.Contains(text, "const HelperConst = 42") {
		t.Errorf("go_header body should land in types.go:\n%s", text)
	}
}

func TestEmitNative_DedupesIdenticalBlocks(t *testing.T) {
	// Two identical go blocks should only contribute one emission.
	mod := compileIvySource(t, `
<<<
go
func dedupedHelper() {}
>>>
<<<
go
func dedupedHelper() {}
>>>
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["native.go"]
	if got := strings.Count(text, "func dedupedHelper()"); got != 1 {
		t.Errorf("duplicate go blocks should dedup; got %d copies:\n%s", got, text)
	}
}

// --- OPEN 056: in-action native go block tests ----------------------

func TestEmitNativeAction_BodyLandsInsideMethod(t *testing.T) {
	mod := compileIvySource(t, `
action shout = {
<<<
go
fmt.Println("hello from native action")
>>>
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, `fmt.Println("hello from native action")`) {
		t.Errorf("native action body should land inline, got:\n%s", actions)
	}
}

func TestEmitNativeAction_CppTaggedBlockSkipped(t *testing.T) {
	mod := compileIvySource(t, `
action mixed = {
<<<
cpp
std::cout << "skipped" << std::endl;
>>>
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if strings.Contains(actions, "std::cout") {
		t.Errorf("cpp-tagged native action should be skipped, got:\n%s", actions)
	}
}

func TestEmitNative_OracleStubReturnsError(t *testing.T) {
	err := CompareGoVsCpp("/tmp/go", "/tmp/cpp", "args")
	if err == nil {
		t.Fatal("oracle stub should return an error")
	}
	if !strings.Contains(err.Error(), "deferred") {
		t.Errorf("oracle stub error = %q, want substring 'deferred'", err.Error())
	}
}
