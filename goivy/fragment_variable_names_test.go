package goivy

import (
	"fmt"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

func TestFragmentVariableNamesShadowingAndDiagnostic(t *testing.T) {
	s := &UninterpretedSort{Name: "node"}
	outer, _ := NewVariable("X", s)
	shadow, _ := NewVariable("X", s)
	named, _ := NewVariable("V!feu!0", s)
	rhs := &LogicExists{Variables: []*LogicVariable{shadow}, Body: &Eq{T1: shadow, T2: named}}
	input := &ForAll{Variables: []*LogicVariable{outer}, Body: &LogicAnd{Terms: []Expr{rhs, &Eq{T1: outer, T2: named}}}}
	c := &checker{variableNames: &fragmentVariableNames{assumes: []fmlaPair{{fmla: input}, {fmla: input}}, asserts: []fmlaPair{{fmla: input}}}}
	legacy := NewVariableUniqifier(nil)
	for i := 0; i < 3; i++ {
		want := legacy.Uniquify(input)
		got := c.variableNames.uniquify(input)
		f := got.(*ForAll)
		parts := f.Body.(*LogicAnd).Terms
		inner := parts[0].(*LogicExists)
		local := inner.Body.(*Eq)
		outside := parts[1].(*Eq)
		if f.Variables[0] == inner.Variables[0] || local.T1 != inner.Variables[0] || outside.T1 != f.Variables[0] || local.T2 != outside.T2 {
			t.Fatal("wrong binding after shadowing")
		}
		// A second fragment check owns an independent numeric name sequence.
		if i == 0 && f.Variables[0].Name != new(fragmentVariableNames).uniquify(input).(*ForAll).Variables[0].Name {
			t.Fatal("nonlocal name generator")
		}
		if c.variableNames.diagnostic != nil {
			t.Fatal("allocated diagnostic names on successful conversion")
		}
		if i == 2 {
			if Key(c.diagnosticFormula(got)) != Key(want) {
				t.Fatal("diagnostic changed bound or free names")
			}
			if Key(c.originalFormula(outside.T2)) != Key(legacy.Undo(want.(*ForAll).Body.(*LogicAnd).Terms[1].(*Eq).T2)) {
				t.Fatal("lost original variable")
			}
		}
	}
}

func BenchmarkFragmentRepeatedVariableNames(b *testing.B) {
	s := &UninterpretedSort{Name: "node"}
	v, _ := NewVariable("X", s)
	f := &ForAll{Variables: []*LogicVariable{v}, Body: &Eq{T1: v, T2: v}}
	for _, count := range []int{128, 1024, 4096} {
		for _, mode := range []string{"source", "counter"} {
			b.Run(fmt.Sprintf("%s/%d", mode, count), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var uniq func(Expr) Expr
					if mode == "source" {
						uniq = NewVariableUniqifier(nil).Uniquify
					} else {
						uniq = new(fragmentVariableNames).uniquify
					}
					for i := 0; i < count; i++ {
						uniq(f)
					}
				}
			})
		}
	}
}

func TestFragmentCounterPreservesSourceSyntaxAndMacroNames(t *testing.T) {
	if xtracer.Enabled && !xtracer.Suppressed {
		t.Skip("active traces require source spellings")
	}
	s := &UninterpretedSort{Name: "node"}
	x, _ := NewVariable("X", s)
	// The duplicate named variable connects the outer binder and the preserved
	// body of Some in the ordinary source representation.
	some := &LogicSome{Params: []Expr{x}, Fmla: &Eq{T1: x, T2: x}}
	f := &ForAll{Variables: []*LogicVariable{x}, Body: some}
	complexHead := &Apply{Func: x, Terms: []Expr{x}}
	for _, test := range []struct {
		label    string
		f, macro Expr
		scalar   bool
	}{
		{"plain", x, nil, true}, {"private-looking variable", &LogicVariable{Name: fragmentVariablePrefix + "0", VSort: s}, nil, true},
		{"some body", f, nil, false}, {"raw binder", &RawForAll{Variables: []*LogicVariable{x}, Body: x}, nil, false},
		{"variable head", complexHead, nil, false}, {"constant macro", x, &IvyDefinition{Lhs: NewConst("formal", s), Rhs: x}, true},
		{"uppercase macro alias", x, &IvyDefinition{Lhs: NewConst("X", s), Rhs: x}, false},
		{"generated-looking constant", NewConst(fragmentVariablePrefix+"0", s), &IvyDefinition{Lhs: NewConst("formal", s), Rhs: x}, false},
	} {
		t.Run(test.label, func(t *testing.T) {
			var macros []fmlaPair
			if test.macro != nil {
				macros = []fmlaPair{{fmla: test.macro}}
			}
			c := new(checker)
			a, _, ms := c.renameFragmentInputs([]fmlaPair{{fmla: test.f}}, nil, macros)
			if (c.variableNames != nil) != test.scalar {
				t.Fatal("wrong name allocator")
			}
			ref := NewVariableUniqifier(nil)
			want := ref.Uniquify(test.f)
			if Key(c.diagnosticFormula(a[0].fmla)) != Key(want) {
				t.Fatalf("assumption differs from source: %s, %s", a[0].fmla, want)
			}
			if len(ms) > 0 {
				want = ref.Uniquify(test.macro)
				if Key(c.diagnosticFormula(ms[0].fmla)) != Key(want) {
					t.Fatal("macro differs from source")
				}
			}
		})
	}
}
