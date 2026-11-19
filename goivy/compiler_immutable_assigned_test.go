package goivy

import (
	"strings"
	"testing"
)

// Python get_symbol_dependencies (ivy_compiler.py) fills an ordered dict
// depth-first: for s in used_symbols_ast(t): res[s] = None, then recurse
// into s's definition. For axiom derived & c with derived := b & a the
// order is derived, b, a, c. goivy used Go maps (random order).
func TestGetSymbolDependenciesFirstOccurrenceOrder(t *testing.T) {
	a, b, c, derived := NewConst("a", Boolean), NewConst("b", Boolean), NewConst("c", Boolean), NewConst("derived", Boolean)
	defMap := map[NodeKey]interface{}{Key(derived): &LogicAnd{Terms: []Expr{b, a}}}
	for i := 0; i < 20; i++ {
		deps := NewInsMap[NodeKey, bool]()
		GetSymbolDependencies(defMap, deps, &LogicAnd{Terms: []Expr{derived, c}})
		var got []NodeKey
		for k := range deps.All() {
			got = append(got, k)
		}
		want := []NodeKey{Key(derived), Key(b), Key(a), Key(c)}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("run %d: got %v, want derived b a c", i, got)
			}
		}
	}
}

// Python (ivy_compiler.py, action interference check):
//
//	raise IvyError(side_effects[s],'immutable symbol assigned. \n{} info: symbol is used in axiom here'.format(lf.lineno))
//	raise IvyError(side_effects[s],'immutable symbol assigned. \n{} Symbol is defined here'.format(lf.lineno))
//
// side_effects[s] is the first subaction modifying s: the error is located
// there and names the axiom's (definition's) location.
func TestCheckImmutableAssignedMatchesPython(t *testing.T) {
	m := mkModuleWithSig()
	s := NewConst("s", Boolean)
	acfg := NewAstConfig()
	axiom := acfg.NewLabeledFormula(acfg.NewAtom("ax"), s)
	axiom.Base.Loc = Location{Filename: "f.ivy", Line: 3}
	m.LabeledAxioms = []*LabeledFormula{axiom}
	assign := NewAssignAction(s, True)
	assign.SetLineno(Location{Filename: "f.ivy", Line: 9})
	side := NewInsMap[NodeKey, ActionsAction]()
	side.Set(Key(s), assign)

	err := checkImmutableAssigned(m, side, map[NodeKey]interface{}{}, false)
	want := "f.ivy: line 9: error: immutable symbol assigned. \nf.ivy: line 3:  info: symbol is used in axiom here"
	if err == nil || err.Error() != want {
		t.Fatalf("axiom: got %v\nwant %s", err, want)
	}

	m.LabeledAxioms = nil
	def := acfg.NewLabeledFormula(acfg.NewAtom("d"), NewDefinition(s, True))
	def.Base.Loc = Location{Filename: "f.ivy", Line: 5}
	m.Definitions = []*LabeledFormula{def}
	err = checkImmutableAssigned(m, side, map[NodeKey]interface{}{}, false)
	if err == nil || !strings.HasSuffix(err.Error(), "immutable symbol assigned. \nf.ivy: line 5:  Symbol is defined here") || !strings.HasPrefix(err.Error(), "f.ivy: line 9: ") {
		t.Fatalf("definition: got %v", err)
	}
}
