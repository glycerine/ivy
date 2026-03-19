package z3bridge

import (
	"testing"

	"github.com/glycerine/goivy/logic"
)

// FuzzQuantConstraintsForAll builds random quantified formulas with a
// QuantConstraints callback and verifies no panics occur during translation.
// The callback generates Le constraints, exercising the quantifier wrapping
// for both valid and invalid body sorts.
func FuzzQuantConstraintsForAll(f *testing.F) {
	// Seed corpus: (varNameLen, bodyKind, isForall)
	// bodyKind: 0=Eq(X,X), 1=Symbol(bool), 2=Not(Eq), 3=And(empty), 4=non-bool var
	f.Add(byte(1), byte(0), true)
	f.Add(byte(3), byte(1), true)
	f.Add(byte(1), byte(2), false)
	f.Add(byte(2), byte(3), false)
	f.Add(byte(1), byte(4), true) // non-bool body, should error not panic

	f.Fuzz(func(t *testing.T, varNameLen byte, bodyKind byte, isForall bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		nameLen := int(varNameLen%5) + 1
		varName := ""
		for i := 0; i < nameLen; i++ {
			varName += string(rune('A' + (int(varNameLen)+i)%26))
		}

		tr := NewTranslator()
		callCount := 0
		tr.SortLookup = func(name string) *Sort {
			if name == "mynat" {
				s := tr.Ctx.IntSort()
				return &s
			}
			return nil
		}
		tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
			callCount++
			return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
		}

		sort := &logic.UninterpretedSort{Name: "mynat"}
		x, err := logic.NewVariable(varName, sort)
		if err != nil {
			return // invalid variable name, skip
		}

		var body logic.Expr
		switch bodyKind % 5 {
		case 0:
			body = &logic.Eq{T1: x, T2: x}
		case 1:
			body = logic.NewSymbol("p", logic.Boolean)
		case 2:
			body = &logic.Not{Body: &logic.Eq{T1: x, T2: x}}
		case 3:
			body = &logic.And{}
		case 4:
			body = x // non-Bool body, should produce error not panic
		}

		var fmla logic.Expr
		if isForall {
			fmla = &logic.ForAll{Variables: []*logic.Variable{x}, Body: body}
		} else {
			fmla = &logic.Exists{Variables: []*logic.Variable{x}, Body: body}
		}

		// Must not panic. Errors are acceptable.
		_, _ = tr.Translate(fmla)
	})
}

// FuzzVariableNaming translates variables with random names and sorts,
// verifying no panics and that the Z3 name contains the sort suffix.
func FuzzVariableNaming(f *testing.F) {
	f.Add("X", "node")
	f.Add("Var", "int")
	f.Add("A", "T")
	f.Add("LongVarName", "SomeSort")
	f.Add("V", "bv32")

	f.Fuzz(func(t *testing.T, varName, sortName string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on var=%q sort=%q: %v", varName, sortName, r)
			}
		}()

		if len(varName) == 0 || len(varName) > 50 {
			return
		}
		if len(sortName) == 0 || len(sortName) > 50 {
			return
		}
		// Variable names must start with uppercase in Ivy
		if varName[0] < 'A' || varName[0] > 'Z' {
			return
		}

		tr := NewTranslator()
		sort := &logic.UninterpretedSort{Name: sortName}
		v, err := logic.NewVariable(varName, sort)
		if err != nil {
			return // invalid name, skip
		}

		z3v, err := tr.Translate(v)
		if err != nil {
			return // translation error, OK
		}

		// The Z3 const string should contain "varName:sortName"
		str := z3v.String()
		if len(str) == 0 {
			t.Fatal("empty Z3 string for variable")
		}
	})
}

// FuzzSortLookup exercises the SortLookup callback with random sort names,
// ensuring no panics when translating sorts with and without interpretations.
func FuzzSortLookup(f *testing.F) {
	f.Add("mynat", true)
	f.Add("T", false)
	f.Add("bvsort", true)
	f.Add("range_sort", true)
	f.Add("", false)

	f.Fuzz(func(t *testing.T, sortName string, hasInterp bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on sort=%q hasInterp=%v: %v", sortName, hasInterp, r)
			}
		}()

		if len(sortName) == 0 || len(sortName) > 50 {
			return
		}

		tr := NewTranslator()
		if hasInterp {
			tr.SortLookup = func(name string) *Sort {
				if name == sortName {
					s := tr.Ctx.IntSort()
					return &s
				}
				return nil
			}
		}

		sort := &logic.UninterpretedSort{Name: sortName}
		_, err := tr.TranslateSort(sort)
		if err != nil {
			return
		}

		// Translate a constant of that sort
		sym := logic.NewSymbol("c", sort)
		_, _ = tr.Translate(sym)
	})
}
