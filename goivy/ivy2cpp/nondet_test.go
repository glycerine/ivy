package ivy2cpp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func TestMkNondetSymStructWithDestructorFields(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
type byte
interpret byte -> bv[8]
type cell
destructor shade(C:cell) : color
destructor bits(C:cell) : byte
action step = {
}
export step
`
	mod := compileIvySource(t, src)
	cell, ok := mod.Sig.Sorts.Get2("cell")
	if !ok {
		t.Fatal("missing cell sort")
	}
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("c", cell)
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, goivy.NewSequence()))
	out, err := Generate(mod, Config{ClassName: "nondetstruct"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"cell c;",
		`c.shade = (color)___ivy_choose(0, "c"`,
		`c.bits = (unsigned)___ivy_choose(0, "c"`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in nondet struct output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestMkNondetSymBitvectorDomain(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type byte
interpret byte -> bv[8]
action step = {
}
export step
`)
	byteSort, ok := mod.Sig.Sorts.Get2("byte")
	if !ok {
		t.Fatal("missing byte sort")
	}
	fnSort, err := goivy.NewFunctionSort(byteSort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("marked", fnSort)
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, goivy.NewSequence()))
	out, err := Generate(mod, Config{ClassName: "nondetbvdom"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"bool marked[256];",
		"for (unsigned X0 = 0; X0 < 256; X0++) {",
		`marked[X0] = (bool)___ivy_choose(0, "marked"`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in bitvector-domain nondet output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestMkNondetSymVariantLocal(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
variant c of t
action step = {
}
export step
`)
	super, ok := mod.Sig.Sorts.Get2("t")
	if !ok {
		t.Fatal("missing t sort")
	}
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("choice", super)
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, goivy.NewSequence()))
	out, err := Generate(mod, Config{ClassName: "nondetvariant"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, terms := range [][]string{
		{"int __ivy_variant", `___ivy_choose(3, "choice"`},
		{"choice = t(0, new t::twrap<a>("},
		{"choice = t(1, new t::twrap<b>("},
		{"choice = t(2, new t::twrap<c>("},
		{`= (a)___ivy_choose(0, "choice"`},
		{`= (b)___ivy_choose(0, "choice"`},
		{`= (c)___ivy_choose(0, "choice"`},
	} {
		if !hasLineWithAllTerms(out.Impl, terms...) {
			t.Fatalf("missing line with terms %v in variant nondet output:\n%s", terms, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestMkNondetSymUninterpretedDomainUsesThunk(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
action step = {
}
export step
`)
	node, ok := mod.Sig.Sorts.Get2("node")
	if !ok {
		t.Fatal("missing node sort")
	}
	fnSort, err := goivy.NewFunctionSort(node, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("marked", fnSort)
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, goivy.NewSequence()))
	out, err := Generate(mod, Config{ClassName: "nondetuninterp"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"hash_thunk<int,bool> marked;",
		"struct __thunk__0 : thunk<int, bool>",
		"marked = hash_thunk<int, bool>(new __thunk__0());",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in uninterpreted-domain nondet output:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "for (int X0") {
		t.Fatalf("uninterpreted domain should not use a bounded iterator:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestMkNondetSymBoundaryWidthsBitVectorLoopHeaders(t *testing.T) {
	for _, tc := range []struct {
		bits int
		card string
		typ  string
	}{
		{1, "2", "unsigned"},
		{2, "4", "unsigned"},
		{8, "256", "unsigned"},
		{16, "65536", "unsigned"},
		{32, "4294967296", "unsigned"},
	} {
		t.Run(fmt.Sprintf("bv%d", tc.bits), func(t *testing.T) {
			mod := compileIvySource(t, fmt.Sprintf(`#lang ivy1.7
type word
interpret word -> bv[%d]
`, tc.bits))
			word, ok := mod.Sig.Sorts.Get2("word")
			if !ok {
				t.Fatal("missing word sort")
			}
			g := &Generator{Mod: mod, ClassName: "nondetbounds"}
			header, err := g.loopHeaderForSort(word, "X0")
			if err != nil {
				t.Fatalf("loopHeaderForSort: %v", err)
			}
			want := fmt.Sprintf("for (%s X0 = 0; X0 < %s; X0++) {", tc.typ, tc.card)
			if header != want {
				t.Fatalf("loop header mismatch:\nwant: %s\n got: %s", want, header)
			}
		})
	}
}

func TestMkNondetSymNoUnsupportedOnLegalInputs(t *testing.T) {
	tests := []struct {
		name string
		src  string
		edit func(t *testing.T, mod *goivy.Module)
	}{
		{
			name: "enum scalar",
			src: `#lang ivy1.7
type color = {red, green}
action step = {}
export step
`,
			edit: func(t *testing.T, mod *goivy.Module) {
				color := mod.Sig.Sorts.Get("color")
				mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", goivy.NewConst("x", color), goivy.NewSequence()))
			},
		},
		{
			name: "bv domain",
			src: `#lang ivy1.7
type byte
interpret byte -> bv[8]
action step = {}
export step
`,
			edit: func(t *testing.T, mod *goivy.Module) {
				byteSort := mod.Sig.Sorts.Get("byte")
				fnSort, err := goivy.NewFunctionSort(byteSort, goivy.Boolean)
				if err != nil {
					t.Fatalf("NewFunctionSort: %v", err)
				}
				mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", goivy.NewConst("marked", fnSort), goivy.NewSequence()))
			},
		},
		{
			name: "variant scalar",
			src: `#lang ivy1.7
type t
variant a of t
variant b of t
action step = {}
export step
`,
			edit: func(t *testing.T, mod *goivy.Module) {
				super := mod.Sig.Sorts.Get("t")
				mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", goivy.NewConst("choice", super), goivy.NewSequence()))
			},
		},
		{
			name: "uninterpreted domain",
			src: `#lang ivy1.7
type node
action step = {}
export step
`,
			edit: func(t *testing.T, mod *goivy.Module) {
				node := mod.Sig.Sorts.Get("node")
				fnSort, err := goivy.NewFunctionSort(node, goivy.Boolean)
				if err != nil {
					t.Fatalf("NewFunctionSort: %v", err)
				}
				mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", goivy.NewConst("marked", fnSort), goivy.NewSequence()))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := compileIvySource(t, tt.src)
			tt.edit(t, mod)
			out, err := Generate(mod, Config{ClassName: "nondetlegal"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			assertNoUnsupportedCPP(t, out)
		})
	}
}
