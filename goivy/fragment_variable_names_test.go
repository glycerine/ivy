package goivy

import (
	"fmt"
	"testing"
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
