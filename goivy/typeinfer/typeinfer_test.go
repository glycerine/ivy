package typeinfer

import (
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

func TestFindBasic(t *testing.T) {
	sv := NewSortVar()
	result := Find(sv)
	if result != sv {
		t.Error("Find on unbound SortVar should return itself")
	}

	s := Wrap(logic.Boolean)
	result = Find(s)
	if result != s {
		t.Error("Find on concrete sort should return itself")
	}
}

func TestFindPathCompression(t *testing.T) {
	sv1 := NewSortVar()
	sv2 := NewSortVar()
	sv3 := NewSortVar()
	s := Wrap(logic.Boolean)

	sv1.Instance = sv2
	sv2.Instance = sv3
	sv3.Instance = s

	result := Find(sv1)
	if result != s {
		t.Error("Find should resolve chain to Boolean")
	}
	// Path compression: sv1 should now point directly to s
	if sv1.Instance != s {
		t.Error("Path compression should update sv1.Instance")
	}
}

func TestUnifySameSorts(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	if err := Unify(Wrap(S), Wrap(S)); err != nil {
		t.Errorf("Same sorts should unify: %v", err)
	}
}

func TestUnifyTopSort(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	if err := Unify(Wrap(logic.TopS), Wrap(S)); err != nil {
		t.Errorf("TopSort should unify with anything: %v", err)
	}
	if err := Unify(Wrap(S), Wrap(logic.TopS)); err != nil {
		t.Errorf("Anything should unify with TopSort: %v", err)
	}
}

func TestUnifySortVar(t *testing.T) {
	sv := NewSortVar()
	S := &logic.UninterpretedSort{Name: "S"}
	if err := Unify(sv, Wrap(S)); err != nil {
		t.Fatalf("SortVar should unify: %v", err)
	}
	result := Find(sv)
	if sw, ok := result.(*SortWrapper); !ok || sw.Sort.String() != "S" {
		t.Error("SortVar should resolve to S after unification")
	}
}

func TestUnifyFunctionSort(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs1, _ := logic.NewFunctionSort(S, S, logic.Boolean)
	fs2, _ := logic.NewFunctionSort(S, S, logic.Boolean)
	if err := Unify(Wrap(fs1), Wrap(fs2)); err != nil {
		t.Errorf("Same FunctionSorts should unify: %v", err)
	}
}

func TestUnifyIncompatible(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	T := &logic.UninterpretedSort{Name: "T"}
	if err := Unify(Wrap(S), Wrap(T)); err == nil {
		t.Error("Different sorts should not unify")
	}
}

func TestUnifyFunctionSortArityMismatch(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs1, _ := logic.NewFunctionSort(S, logic.Boolean)
	fs2, _ := logic.NewFunctionSort(S, S, logic.Boolean)
	if err := Unify(Wrap(fs1), Wrap(fs2)); err == nil {
		t.Error("FunctionSorts with different arity should not unify")
	}
}

func TestOccursIn(t *testing.T) {
	sv := NewSortVar()
	S := &logic.UninterpretedSort{Name: "S"}

	if OccursIn(sv, Wrap(S)) {
		t.Error("SortVar should not occur in unrelated sort")
	}
	if !OccursIn(sv, sv) {
		t.Error("SortVar should occur in itself")
	}
}

func TestConvertFromSortVars(t *testing.T) {
	sv := NewSortVar()
	result := ConvertFromSortVars(sv)
	if _, ok := result.(*logic.TopSort); !ok {
		t.Error("Unbound SortVar should convert to TopSort")
	}

	S := &logic.UninterpretedSort{Name: "S"}
	sv2 := NewSortVar()
	sv2.Instance = Wrap(S)
	result2 := ConvertFromSortVars(sv2)
	if result2.String() != "S" {
		t.Errorf("Bound SortVar should convert to S, got %s", result2)
	}
}

func TestConvertToSortVars(t *testing.T) {
	result := ConvertToSortVars(logic.TopS)
	if _, ok := result.(*SortVar); !ok {
		t.Error("TopSort should convert to SortVar")
	}

	S := &logic.UninterpretedSort{Name: "S"}
	result2 := ConvertToSortVars(S)
	if sw, ok := result2.(*SortWrapper); !ok || sw.Sort.String() != "S" {
		t.Error("Non-TopSort should wrap to SortWrapper")
	}
}

// Reproduce type_inference.py __main__ examples
func TestConcretizeSortsBasic(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := logic.NewFunctionSort(S, logic.Boolean)

	XS, _ := logic.NewVariable("X", S)
	xs := logic.NewConst("x", S)
	ps := logic.NewConst("p", UnaryRelationS)

	// f1 = And(ps(XS), ps(xs))
	psXS, err := logic.NewApply(ps, XS)
	if err != nil {
		t.Fatal(err)
	}
	psxs, err := logic.NewApply(ps, xs)
	if err != nil {
		t.Fatal(err)
	}
	f1, err := logic.NewAnd(psXS, psxs)
	if err != nil {
		t.Fatal(err)
	}

	cf1, err := ConcretizeSorts(f1, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("f1:", f1)
	t.Log("cf1:", cf1)

	// cf1 should equal f1 since sorts are already concrete
	if !f1.Equal(cf1) {
		t.Error("Concretized fully-typed term should equal original")
	}
}

func TestConcretizeSortsWithTopSort(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := logic.NewFunctionSort(S, logic.Boolean)
	UnaryRelationT, _ := logic.NewFunctionSort(logic.TopS, logic.Boolean)

	XT, _ := logic.NewVariable("X", logic.TopS)
	xs := logic.NewConst("x", S)
	ps := logic.NewConst("p", UnaryRelationS)
	pt := logic.NewConst("p", UnaryRelationT)

	// f2 = And(ps(XT), pt(xs))
	psXT, err := logic.NewApply(ps, XT)
	if err != nil {
		t.Fatal(err)
	}
	ptxs, err := logic.NewApply(pt, xs)
	if err != nil {
		t.Fatal(err)
	}
	f2, err := logic.NewAnd(psXT, ptxs)
	if err != nil {
		t.Fatal(err)
	}

	cf2, err := ConcretizeSorts(f2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("f2:", f2)
	t.Log("cf2:", cf2)
	// After concretization, all TopSorts should be resolved to S
}

func TestConcretizeNamedBinder(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := logic.NewFunctionSort(S, logic.Boolean)

	XT, _ := logic.NewVariable("X", logic.TopS)
	ps := logic.NewConst("p", UnaryRelationS)
	psXT, _ := logic.NewApply(ps, XT)

	f6, err := logic.NewNamedBinder("mybinder", []*logic.Variable{XT}, nil, psXT)
	if err != nil {
		t.Fatal(err)
	}

	cf6, err := ConcretizeSorts(f6, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("f6:", f6)
	t.Log("f6.sort:", f6.NodeSort())
	t.Log("cf6:", cf6)
	t.Log("cf6.sort:", cf6.NodeSort())
}

func FuzzUnify(f *testing.F) {
	f.Add(true, true)
	f.Add(true, false)
	f.Add(false, false)
	f.Fuzz(func(t *testing.T, useVar1, useVar2 bool) {
		var s1, s2 SortOrVar
		if useVar1 {
			s1 = NewSortVar()
		} else {
			s1 = Wrap(&logic.UninterpretedSort{Name: "S"})
		}
		if useVar2 {
			s2 = NewSortVar()
		} else {
			s2 = Wrap(&logic.UninterpretedSort{Name: "S"})
		}
		// Should not panic
		_ = Unify(s1, s2)
	})
}
