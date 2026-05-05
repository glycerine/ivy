package goivy

import (
	"testing"
)

func TestFindBasic(t *testing.T) {
	sv := NewSortVar()
	result := TypeInferFind(sv)
	if result != sv {
		t.Error("Find on unbound SortVar should return itself")
	}

	s := Wrap(Boolean)
	result = TypeInferFind(s)
	if result != s {
		t.Error("Find on concrete sort should return itself")
	}
}

func TestFindPathCompression(t *testing.T) {
	sv1 := NewSortVar()
	sv2 := NewSortVar()
	sv3 := NewSortVar()
	s := Wrap(Boolean)

	sv1.Instance = sv2
	sv2.Instance = sv3
	sv3.Instance = s

	result := TypeInferFind(sv1)
	if result != s {
		t.Error("Find should resolve chain to Boolean")
	}
	// Path compression: sv1 should now point directly to s
	if sv1.Instance != s {
		t.Error("Path compression should update sv1.Instance")
	}
}

func TestUnifySameSorts(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	if err := TypeInferUnify(Wrap(S), Wrap(S)); err != nil {
		t.Errorf("Same sorts should unify: %v", err)
	}
}

func TestUnifyTopSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	if err := TypeInferUnify(Wrap(TopS), Wrap(S)); err != nil {
		t.Errorf("TopSort should unify with anything: %v", err)
	}
	if err := TypeInferUnify(Wrap(S), Wrap(TopS)); err != nil {
		t.Errorf("Anything should unify with TopSort: %v", err)
	}
}

func TestUnifySortVar(t *testing.T) {
	sv := NewSortVar()
	S := &UninterpretedSort{Name: "S"}
	if err := TypeInferUnify(sv, Wrap(S)); err != nil {
		t.Fatalf("SortVar should unify: %v", err)
	}
	result := TypeInferFind(sv)
	if sw, ok := result.(*SortWrapper); !ok || sw.Sort.String() != "S" {
		t.Error("SortVar should resolve to S after unification")
	}
}

func TestUnifyFunctionSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs1, _ := NewFunctionSort(S, S, Boolean)
	fs2, _ := NewFunctionSort(S, S, Boolean)
	if err := TypeInferUnify(Wrap(fs1), Wrap(fs2)); err != nil {
		t.Errorf("Same FunctionSorts should unify: %v", err)
	}
}

func TestUnifyIncompatible(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	T := &UninterpretedSort{Name: "T"}
	if err := TypeInferUnify(Wrap(S), Wrap(T)); err == nil {
		t.Error("Different sorts should not unify")
	}
}

func TestUnifyFunctionSortArityMismatch(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs1, _ := NewFunctionSort(S, Boolean)
	fs2, _ := NewFunctionSort(S, S, Boolean)
	if err := TypeInferUnify(Wrap(fs1), Wrap(fs2)); err == nil {
		t.Error("FunctionSorts with different arity should not unify")
	}
}

func TestOccursIn(t *testing.T) {
	sv := NewSortVar()
	S := &UninterpretedSort{Name: "S"}

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
	if _, ok := result.(*TopSort); !ok {
		t.Error("Unbound SortVar should convert to TopSort")
	}

	S := &UninterpretedSort{Name: "S"}
	sv2 := NewSortVar()
	sv2.Instance = Wrap(S)
	result2 := ConvertFromSortVars(sv2)
	if result2.String() != "S" {
		t.Errorf("Bound SortVar should convert to S, got %s", result2)
	}
}

func TestConvertToSortVars(t *testing.T) {
	result := ConvertToSortVars(TopS)
	if _, ok := result.(*SortVar); !ok {
		t.Error("TopSort should convert to SortVar")
	}

	S := &UninterpretedSort{Name: "S"}
	result2 := ConvertToSortVars(S)
	if sw, ok := result2.(*SortWrapper); !ok || sw.Sort.String() != "S" {
		t.Error("Non-TopSort should wrap to SortWrapper")
	}
}

// Reproduce type_inference.py __main__ examples
func TestConcretizeSortsBasic(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := NewFunctionSort(S, Boolean)

	XS, _ := NewVariable("X", S)
	xs := NewConst("x", S)
	ps := NewConst("p", UnaryRelationS)

	// f1 = And(ps(XS), ps(xs))
	psXS, err := NewApply(ps, XS)
	if err != nil {
		t.Fatal(err)
	}
	psxs, err := NewApply(ps, xs)
	if err != nil {
		t.Fatal(err)
	}
	f1, err := NewAnd(psXS, psxs)
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
	S := &UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := NewFunctionSort(S, Boolean)
	UnaryRelationT, _ := NewFunctionSort(TopS, Boolean)

	XT, _ := NewVariable("X", TopS)
	xs := NewConst("x", S)
	ps := NewConst("p", UnaryRelationS)
	pt := NewConst("p", UnaryRelationT)

	// f2 = And(ps(XT), pt(xs))
	psXT, err := NewApply(ps, XT)
	if err != nil {
		t.Fatal(err)
	}
	ptxs, err := NewApply(pt, xs)
	if err != nil {
		t.Fatal(err)
	}
	f2, err := NewAnd(psXT, ptxs)
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
	S := &UninterpretedSort{Name: "S"}
	UnaryRelationS, _ := NewFunctionSort(S, Boolean)

	XT, _ := NewVariable("X", TopS)
	ps := NewConst("p", UnaryRelationS)
	psXT, _ := NewApply(ps, XT)

	f6, err := NewNamedBinder("mybinder", []*Variable{XT}, nil, psXT)
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
			s1 = Wrap(&UninterpretedSort{Name: "S"})
		}
		if useVar2 {
			s2 = NewSortVar()
		} else {
			s2 = Wrap(&UninterpretedSort{Name: "S"})
		}
		// Should not panic
		_ = TypeInferUnify(s1, s2)
	})
}
