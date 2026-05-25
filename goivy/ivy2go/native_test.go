package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
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

// --- OPEN 064: antiquote substitution tests -------------------------

func TestRenderNativeGoTemplate_BadIndexLeavesMarker(t *testing.T) {
	g := newExprGen(t, "")
	got := g.renderNativeGoTemplate("fmt.Println(`xyz`)", nil)
	if !strings.Contains(got, "bad antiquote") {
		t.Errorf("non-numeric antiquote should leave a marker, got: %q", got)
	}
}

func TestRenderNativeGoTemplate_OutOfRangeIndex(t *testing.T) {
	g := newExprGen(t, "")
	got := g.renderNativeGoTemplate("`5`", nil)
	if !strings.Contains(got, "out of range") {
		t.Errorf("out-of-range index should leave a marker, got: %q", got)
	}
}

func TestRenderNativeGoTemplate_PassthroughWhenNoBackticks(t *testing.T) {
	g := newExprGen(t, "")
	got := g.renderNativeGoTemplate("fmt.Println(\"hi\")", nil)
	if got != `fmt.Println("hi")` {
		t.Errorf("plain body should pass through, got: %q", got)
	}
}

func TestRenderNativeGoTemplate_SubstitutesParam(t *testing.T) {
	g := newExprGen(t, "")
	p := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got := g.renderNativeGoTemplate("v := `0`", []goivy.Expr{p})
	if got != "v := true" {
		t.Errorf("antiquote should substitute param, got: %q", got)
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

func TestCompareGoVsCpp_NonexistentBinariesFailWithRunError(t *testing.T) {
	err := CompareGoVsCpp("/tmp/does-not-exist", "/tmp/does-not-exist", "")
	if err == nil {
		t.Fatal("oracle should error on missing binaries")
	}
	if !strings.Contains(err.Error(), "go binary") {
		t.Errorf("oracle error should mention go binary, got: %v", err)
	}
}
