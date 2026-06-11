package ivyrepl

import (
	"testing"

	"github.com/glycerine/ivy/goivy/zlisp"
)

func TestZ3RawSmoke(t *testing.T) {
	env := NewEnv(false)
	defer env.Close()

	res, err := env.EvalString(`
(def ctx (z3_new_context))
(def intsort (z3_int_sort ctx))
(def x (z3_const ctx "x" intsort))
(def solver (z3_new_solver ctx))
(z3_assert solver (z3_gt ctx x (z3_int ctx 2)))
(z3_check solver)
`)
	if err != nil {
		t.Fatalf("raw z3 repl evaluation failed: %v", err)
	}
	str, ok := res.(*zlisp.SexpStr)
	if !ok {
		t.Fatalf("z3_check returned %T, want string", res)
	}
	if str.S != "sat" {
		t.Fatalf("z3_check = %q, want sat", str.S)
	}
}

func TestIvySolverSmoke(t *testing.T) {
	env := NewEnv(false)
	defer env.Close()

	res, err := env.EvalString(`
(def solver (ivy_new_solver))
(def formula (ivy_parse_formula "true"))
(ivy_is_sat solver formula)
`)
	if err != nil {
		t.Fatalf("ivy solver repl evaluation failed: %v", err)
	}
	b, ok := res.(*zlisp.SexpBool)
	if !ok {
		t.Fatalf("ivy_is_sat returned %T, want bool", res)
	}
	if !b.Val {
		t.Fatal("ivy_is_sat true = false, want true")
	}

	res, err = env.EvalString(`
(def solver2 (ivy_new_solver))
(def premise (ivy_true))
(def conclusion (ivy_parse_formula "true"))
(ivy_implies solver2 premise conclusion)
`)
	if err != nil {
		t.Fatalf("ivy implies repl evaluation failed: %v", err)
	}
	b, ok = res.(*zlisp.SexpBool)
	if !ok {
		t.Fatalf("ivy_implies returned %T, want bool", res)
	}
	if !b.Val {
		t.Fatal("ivy_implies true true = false, want true")
	}
}
