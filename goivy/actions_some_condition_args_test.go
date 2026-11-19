package goivy

import "testing"

// An ivy_ast.Some keeps [*params, fmla] (plus the index for SomeMin and
// SomeMax) in args, so substitute_constants_ast traces it as
// "type=Some nargs=2" and substitutes inside it (test/crashpong2.ivy).
func TestSomeConditionArgsLikePythonSome(t *testing.T) {
	x := NewConst("self:q", Boolean)
	c := NewConst("c", Boolean)
	d := NewConst("d", Boolean)
	fm, err := NewApply(NewConst("r", LogicRelationSort([]Sort{Boolean, Boolean})), c, x)
	if err != nil {
		t.Fatal(err)
	}
	sc := &SomeCondition{Params: []*Const{x}, Fmla: fm, Kind: "some"}
	if got := TypeName(sc); got != "Some" {
		t.Fatalf("type name %q", got)
	}
	if n := len(sc.Args()); n != 2 {
		t.Fatalf("nargs %d", n)
	}
	out, ok := SubstituteConstantsAST(sc, map[NodeKey]Expr{Key(c): d}).(*SomeCondition)
	if !ok {
		t.Fatalf("substitution result %T", out)
	}
	if got := out.String(); got != "some self:q. r(d,self:q)" {
		t.Fatalf("substituted %q", got)
	}
	if got := sc.String(); got != "some self:q. r(c,self:q)" {
		t.Fatalf("original changed: %q", got)
	}
}
