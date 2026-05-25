package ivy2go

import (
	"strings"
	"testing"
)

// --- OPEN 063: variant constructors + *> downcast tests --------------

func TestEmitVariant_SuperStructAndConstructor(t *testing.T) {
	mod := compileIvySource(t, `
type t0
type t1
variant t1 of t0
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if text == "" {
		t.Fatalf("types.go missing:\n%v", keysOf(out.Files))
	}
	// Super struct with Tag field.
	if !strings.Contains(text, "type T0 struct") {
		t.Errorf("super struct missing, got:\n%s", text)
	}
	if !strings.Contains(text, "Tag int") {
		t.Errorf("Tag field missing, got:\n%s", text)
	}
	// Plain-leaf constructor (no payload arg).
	if !strings.Contains(text, "func NewT0T1() T0 { return T0{Tag: 0} }") {
		t.Errorf("plain-leaf constructor missing, got:\n%s", text)
	}
}

func TestEmitVariant_TwoLeafConstructors(t *testing.T) {
	mod := compileIvySource(t, `
type super
type leaf_a
type leaf_b
variant leaf_a of super
variant leaf_b of super
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "NewSuperLeafA") {
		t.Errorf("missing NewSuperLeafA constructor, got:\n%s", text)
	}
	if !strings.Contains(text, "NewSuperLeafB") {
		t.Errorf("missing NewSuperLeafB constructor, got:\n%s", text)
	}
}
