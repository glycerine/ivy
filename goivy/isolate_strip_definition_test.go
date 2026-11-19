package goivy

import "testing"

// Python strip_labeled_fmla (ivy_isolate.py) collects the strip binding
// with get_strip_binding and then runs strip_action over the formula.
// strip_action recurses through every node (args = [strip_action(arg)...];
// ...; return ast.clone(args)), so an application nested inside a
// Definition, gap(X,k,m) = ..., loses its leading isolate parameter and
// its sort loses the matching domain sort: gap(k,m). Before the fix,
// goivy's stripNode returned a Definition unchanged, and the sht.ivy
// isolate iso_dm(me:id) kept imap.gap(X:id,k,m).
func TestStripLabeledFormulaStripsInsideDefinition(t *testing.T) {
	m := mkModuleWithSig()
	id, it := isolateMkSort("id"), isolateMkSort("key.iter.t")
	gapSort, err := NewFunctionSort(id, it, it, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	const name = "proto.impl.dm.impl.imap.gap"
	m.Sig.Symbols.Set(name, &SymbolEntry{Name: name, Sort: gapSort})

	x, err := NewVariable("X", id)
	if err != nil {
		t.Fatal(err)
	}
	k, mm := NewConst("k", it), NewConst("m", it)
	lhs := MustApply(NewConst(name, gapSort), x, k, mm)
	lf := NewAstConfig().NewLabeledFormula(NewAstConfig().NewAtom("proto.impl.dm.impl.imap"), NewDefinition(lhs, True))

	res := StripLabeledFormula(lf, StripMapOf(StripPair{"proto.impl.dm", []string{"me"}}), m)

	def, ok := res.Formula.(*LogicDefinition)
	if !ok {
		t.Fatalf("formula %T", res.Formula)
	}
	app, ok := def.Lhs.(*Apply)
	if !ok {
		t.Fatalf("lhs %T", def.Lhs)
	}
	if len(app.Terms) != 2 || Key(app.Terms[0]) != Key(k) || Key(app.Terms[1]) != Key(mm) {
		t.Fatalf("lhs = %v, want %s(k,m)", app, name)
	}
	want, err := NewFunctionSort(it, it, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if got := app.Func.(*Const).CSort; Key(NewConst(name, got)) != Key(NewConst(name, want)) {
		t.Fatalf("gap sort = %v, want %v", got, want)
	}
}
