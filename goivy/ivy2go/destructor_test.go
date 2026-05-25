package ivy2go

import (
	"strings"
	"testing"
)

// --- OPEN 053: destructor + variant struct emission tests -----------

func TestEmitDestructor_BasicRecordStruct(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if text == "" {
		t.Fatalf("types.go missing:\n%v", keysOf(out.Files))
	}
	requireHasLineWithAllTerms(t, text, "type Point struct")
	requireHasLineWithAllTerms(t, text, "X bool")
	requireHasLineWithAllTerms(t, text, "Y bool")
}

func TestEmitDestructor_HasEqualMethod(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "func (a Point) Equal(b Point) bool") {
		t.Errorf("Point should have Equal method, got:\n%s", text)
	}
	if !strings.Contains(text, "if a.X != b.X") {
		t.Errorf("Equal should compare X field, got:\n%s", text)
	}
}

func TestEmitDestructor_IndexedFieldUsesArray(t *testing.T) {
	mod := compileIvySource(t, `
type idx = {0..3}
type holder = struct {
    cell(I: idx) : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "Cell [4]bool") {
		t.Errorf("indexed destructor should lower to [4]bool field, got:\n%s", text)
	}
}

func TestEmitDestructorApply_ReadsField(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
relation flag
action set_to_x(p: point) = {
	flag := x(p)
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "s.Flag = p.X") {
		t.Errorf("destructor read should emit p.X, got:\n%s", actions)
	}
}

// --- OPEN 062: Hash + Less methods on destructor structs ------------

func TestEmitDestructor_HasHashMethod(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "func (a Point) Hash() uint64") {
		t.Errorf("Point should have Hash method, got:\n%s", text)
	}
	if !strings.Contains(text, "mixHash(h, a.X)") {
		t.Errorf("Hash should mix in X field, got:\n%s", text)
	}
}

func TestEmitDestructor_HasLessMethod(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "func (a Point) Less(b Point) bool") {
		t.Errorf("Point should have Less method, got:\n%s", text)
	}
	if !strings.Contains(text, "lessOrd(a.X, b.X)") {
		t.Errorf("Less should compare X via lessOrd, got:\n%s", text)
	}
}

// --- OPEN 062.1: deterministic hash for hash-thunk fields -----------

func TestEmitDestructorHash_MapFieldSortsKeys(t *testing.T) {
	// Indexed destructor with a large enough domain to force map
	// storage. cell(I: idx) where idx has 2048 values → map.
	mod := compileIvySource(t, `
type idx = {0..2048}
type holder = struct {
    cell(I: idx) : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "sort.Slice(keys") {
		t.Errorf("map-field Hash should sort keys, got:\n%s", text)
	}
	if !strings.Contains(text, "lessOrd(keys[i], keys[j])") {
		t.Errorf("sort ordering should go through lessOrd, got:\n%s", text)
	}
}

func TestEmitDestructor_RuntimeHelpersEmittedOnDemand(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "func mixHash(h uint64, v any) uint64") {
		t.Errorf("mixHash helper missing, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "func lessOrd(a, b any) bool") {
		t.Errorf("lessOrd helper missing, got:\n%s", runtime)
	}
}

func TestEmitDestructor_AllFilesGofmtClean(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct {
    x : bool,
    y : bool
}
relation flag
action set_to_x(p: point) = {
	flag := x(p)
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
