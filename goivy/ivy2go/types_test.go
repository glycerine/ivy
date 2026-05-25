package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- M2: type emission tests -----------------------------------------

func TestEmitEnumProducesIotaBlock(t *testing.T) {
	mod := compileIvySource(t, "type color = {red, green, blue}")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if text == "" {
		t.Fatalf("types.go is empty:\n%v", out.Files)
	}
	requireHasLineWithAllTerms(t, text, "type", "Color", "int")
	requireHasLineWithAllTerms(t, text, "Red", "Color", "iota")
	requireHasLineWithAllTerms(t, text, "Green")
	requireHasLineWithAllTerms(t, text, "Blue")
}

func TestEmitNumericEnumNotEmittedAsType(t *testing.T) {
	// Numeric-extension enums map directly to plain int per
	// ivy2cpp; no `type X int` declaration should land in types.go.
	mod := compileIvySource(t, `type three = {0, 1, 2}`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	// Should NOT declare `type Three int` since it's numeric.
	if hasLineWithAllTerms(text, "type", "Three") {
		t.Errorf("numeric enum should not produce a named type:\n%s", text)
	}
}

func TestEmitRangeProducesTypedAlias(t *testing.T) {
	mod := compileIvySource(t, `type idx = {0..7}`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	requireHasLineWithAllTerms(t, text, "type", "Idx", "int")
	requireHasLineWithAllTerms(t, text, "Range[0..7]")
}

func TestEmitUninterpretedProducesTypedInt(t *testing.T) {
	mod := compileIvySource(t, `type node`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	requireHasLineWithAllTerms(t, text, "type", "Node", "int")
}

func TestEmitMultipleSortsInDeclarationOrder(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green, blue}
type idx = {0..3}
type node
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	// All three should appear.
	for _, want := range []string{"type Color int", "type Idx int", "type Node int"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in types.go:\n%s", want, text)
		}
	}
	// Color (a non-numeric enum) needs its iota block.
	requireHasLineWithAllTerms(t, text, "Red", "Color", "iota")
}

// --- Tier 3: every emitted file remains gofmt-clean after M2 ---------

func TestEmitTypesFilesAreGofmtClean(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green, blue}
type idx = {0..3}
type node
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

// --- goType / goScalarType unit tests --------------------------------

func TestGoTypeBool(t *testing.T) {
	if got := goType(goivy.Boolean); got != "bool" {
		t.Errorf("goType(Boolean) = %q, want bool", got)
	}
}

func TestGoCardinalType(t *testing.T) {
	if got := goCardinalType(8); got != "uint32" {
		t.Errorf("goCardinalType(8) = %q, want uint32", got)
	}
	// Anything above uint32 max maps to uint64.
	if got := goCardinalType(int(^uint32(0)) + 1); got != "uint64" {
		t.Errorf("goCardinalType(large) = %q, want uint64", got)
	}
}

func TestParseGoInterpType(t *testing.T) {
	// Note: parseBracketInts consumes consecutive [X] segments, not
	// comma-separated lists. So intbv's three params are spelled
	// "intbv[lo][hi][bits]", matching ivy2cpp's parsing rule.
	cases := map[string]goInterpType{
		"bv[16]":          {Kind: goInterpBV, Bits: 16},
		"bv[128]":         {Kind: goInterpBV, Bits: 128},
		"strbv[256]":      {Kind: goInterpStrBV, Bits: 256},
		"intbv[0][10][8]": {Kind: goInterpIntBV, Lo: 0, Hi: 10, Bits: 8},
	}
	for text, want := range cases {
		got, ok := parseGoInterpType(text)
		if !ok {
			t.Errorf("parseGoInterpType(%q) = !ok", text)
			continue
		}
		if got != want {
			t.Errorf("parseGoInterpType(%q) = %+v, want %+v", text, got, want)
		}
	}
}

func TestGoInterpTypePrimitiveBV(t *testing.T) {
	cases := map[int]string{
		8:   "uint32",
		32:  "uint32",
		33:  "uint64",
		64:  "uint64",
		65:  "Uint128",
		128: "Uint128",
		129: "",
	}
	for bits, want := range cases {
		it := goInterpType{Kind: goInterpBV, Bits: bits}
		if got := it.primitiveType(); got != want {
			t.Errorf("primitiveType(bv[%d]) = %q, want %q", bits, got, want)
		}
	}
}

func TestGoFunctionStorageScalarVsArrayVsMap(t *testing.T) {
	// Domain []enum-with-card-3 → array storage (3 ≤ largeThresh).
	colorSort := &goivy.LogicEnumeratedSort{
		Name:      "color",
		Extension: []string{"red", "green", "blue"},
	}
	boolSort := goivy.Boolean
	st := goFunctionStorageFor(nil, []goivy.Sort{colorSort}, boolSort)
	if st.Kind != goStorageArray {
		t.Errorf("colorSort→bool storage = %v, want array", st.Kind)
	}
	if st.Dims[0] != 3 {
		t.Errorf("colorSort→bool dims[0] = %d, want 3", st.Dims[0])
	}
	// Empty domain → scalar.
	st = goFunctionStorageFor(nil, nil, boolSort)
	if st.Kind != goStorageScalar {
		t.Errorf("empty-domain storage = %v, want scalar", st.Kind)
	}
}

func TestGoArrayPrefixOrder(t *testing.T) {
	if got := goArrayPrefix([]int{3, 5}); got != "[3][5]" {
		t.Errorf("goArrayPrefix([3,5]) = %q, want [3][5]", got)
	}
	if got := goArrayPrefix(nil); got != "" {
		t.Errorf("goArrayPrefix(nil) = %q, want empty", got)
	}
}

func TestEmitTypesEmptyModuleProducesNoTypesFile(t *testing.T) {
	// Module with no sorts → types.go is whitespace-only and
	// therefore omitted (per finalize's always=false default).
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, has := out.Files["types.go"]; has {
		t.Errorf("empty module should not produce types.go; got:\n%s", out.Files["types.go"])
	}
}
