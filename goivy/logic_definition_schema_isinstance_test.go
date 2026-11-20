package goivy

import "testing"

// Python DefinitionSchema subclasses Definition: isinstance(x, il.Definition)
// accepts it, and il.equal_mod_alpha compares a schema and a plain
// definition by their args (test/ind1.ivy).
func TestDefinitionSchemaIsADefinition(t *testing.T) {
	s := &UninterpretedSort{Name: "t"}
	f := NewConst("f", FuncConstSort(s, s))
	x := &LogicVariable{Name: "X", VSort: s}
	lhs, _ := NewApply(f, x)
	ds := NewIvyDefinitionSchema(lhs, x)
	d := NewIvyDefinition(lhs, x)
	if _, ok := asDefinition(ds); !ok {
		t.Fatal("DefinitionSchema is not a Definition")
	}
	if !IlEqualModAlpha(ds, d) || !IlEqualModAlpha(d, ds) {
		t.Fatal("schema and definition with equal args differ")
	}
	if err := CheckConcsMatch(NewAstConfig().NewLabeledFormula(nil, ds), NewAstConfig().NewLabeledFormula(nil, d)); err != nil {
		t.Fatal(err)
	}
}
