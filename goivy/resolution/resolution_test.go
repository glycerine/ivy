package resolution

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

// helper to make a variable with TopSort (Python variables have no case restriction in ivy_logic)
func mkVar(name string) *logic.Variable {
	v, err := logic.NewVariable(name, logic.TopS)
	if err != nil {
		panic(fmt.Sprintf("mkVar(%q): %v", name, err))
	}
	return v
}

func mkConst(name string) *logic.Const {
	return logic.NewConst(name, logic.TopS)
}

func envString(env Env) string {
	var parts []string
	for k, v := range env {
		parts = append(parts, fmt.Sprintf("%s->%s", k, v.String()))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// TestEnvFindConstant tests that EnvFind returns a constant unchanged.
func TestEnvFindConstant(t *testing.T) {
	env := make(Env)
	c := mkConst("a")
	result := EnvFind(env, c)
	if result != c {
		t.Errorf("expected constant a, got %v", result)
	}
}

// TestEnvFindUnmappedVar tests that an unmapped variable returns itself.
func TestEnvFindUnmappedVar(t *testing.T) {
	env := make(Env)
	x := mkVar("X")
	result := EnvFind(env, x)
	if result != x {
		t.Errorf("expected var X, got %v", result)
	}
}

// TestEnvFindSingleStep tests one-step variable resolution.
func TestEnvFindSingleStep(t *testing.T) {
	env := make(Env)
	x := mkVar("X")
	a := mkConst("a")
	env["X"] = a
	result := EnvFind(env, x)
	if result != a {
		t.Errorf("expected constant a, got %v", result)
	}
}

// TestEnvFindChain tests multi-step variable chain resolution.
func TestEnvFindChain(t *testing.T) {
	env := make(Env)
	x := mkVar("X")
	y := mkVar("Y")
	a := mkConst("a")
	env["X"] = y
	env["Y"] = a
	result := EnvFind(env, x)
	if result != a {
		t.Errorf("expected constant a, got %v", result)
	}
}

// TestTermsMGUPythonExample mirrors the Python test_terms_mgu.
func TestTermsMGUPythonExample(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	z := mkVar("Z")
	a := mkConst("a")
	b := mkConst("b")

	match, subs := TermsMGU(
		[]ResolutionTerm{x, y, x, b, y},
		[]ResolutionTerm{a, z, a, z, b},
	)
	if !match {
		t.Fatal("expected match=true")
	}
	// Expected: X->a, Y->b, Z->b
	if subs["X"].String() != "a" {
		t.Errorf("X: expected a, got %s", subs["X"])
	}
	if subs["Y"].String() != "b" {
		t.Errorf("Y: expected b, got %s", subs["Y"])
	}
	if subs["Z"].String() != "b" {
		t.Errorf("Z: expected b, got %s", subs["Z"])
	}
}

// TestTermsMGULengthMismatch tests that different-length term lists fail.
func TestTermsMGULengthMismatch(t *testing.T) {
	a := mkConst("a")
	match, _ := TermsMGU([]ResolutionTerm{a}, []ResolutionTerm{a, a})
	if match {
		t.Error("expected match=false for length mismatch")
	}
}

// TestTermsMGUEmptyLists tests that empty term lists unify successfully.
func TestTermsMGUEmptyLists(t *testing.T) {
	match, subs := TermsMGU(nil, nil)
	if !match {
		t.Fatal("expected match=true for empty lists")
	}
	if len(subs) != 0 {
		t.Errorf("expected empty subs, got %v", subs)
	}
}

// TestTermsMGUConstantMismatch tests that mismatched constants fail.
func TestTermsMGUConstantMismatch(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	match, _ := TermsMGU([]ResolutionTerm{a}, []ResolutionTerm{b})
	if match {
		t.Error("expected match=false for constant mismatch")
	}
}

// TestTermsMGUSameConstants tests that identical constants unify with empty subs.
func TestTermsMGUSameConstants(t *testing.T) {
	a1 := mkConst("a")
	a2 := mkConst("a")
	match, subs := TermsMGU([]ResolutionTerm{a1}, []ResolutionTerm{a2})
	if !match {
		t.Fatal("expected match=true")
	}
	if len(subs) != 0 {
		t.Errorf("expected empty subs, got %v", subs)
	}
}

// TestTermsMGUVarToVar tests that two different variables unify.
func TestTermsMGUVarToVar(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	match, subs := TermsMGU([]ResolutionTerm{x}, []ResolutionTerm{y})
	if !match {
		t.Fatal("expected match=true")
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 substitution, got %d", len(subs))
	}
	if subs["X"].String() != "Y" {
		t.Errorf("expected X->Y, got X->%s", subs["X"])
	}
}

// TestTermsMGUSameVar tests that the same variable on both sides unifies trivially.
func TestTermsMGUSameVar(t *testing.T) {
	x := mkVar("X")
	match, subs := TermsMGU([]ResolutionTerm{x}, []ResolutionTerm{x})
	if !match {
		t.Fatal("expected match=true")
	}
	if len(subs) != 0 {
		t.Errorf("expected empty subs, got %v", subs)
	}
}

// TestTermsMGUSortMismatch tests that terms with different sorts fail.
func TestTermsMGUSortMismatch(t *testing.T) {
	s1 := &logic.UninterpretedSort{Name: "S1"}
	s2 := &logic.UninterpretedSort{Name: "S2"}
	c1 := logic.NewConst("a", s1)
	c2 := logic.NewConst("a", s2)
	match, _ := TermsMGU([]ResolutionTerm{c1}, []ResolutionTerm{c2})
	if match {
		t.Error("expected match=false for sort mismatch")
	}
}

// TestMGUSameRelation tests atom unification with same relation name.
func TestMGUSameRelation(t *testing.T) {
	x := mkVar("X")
	a := mkConst("a")
	atom1 := NewAtom("r", x)
	atom2 := NewAtom("r", a)
	match, subs := MGU(atom1, atom2)
	if !match {
		t.Fatal("expected match=true")
	}
	if subs["X"].String() != "a" {
		t.Errorf("X: expected a, got %s", subs["X"])
	}
}

// TestMGUDifferentRelation tests that atoms with different names fail.
func TestMGUDifferentRelation(t *testing.T) {
	a := mkConst("a")
	atom1 := NewAtom("r", a)
	atom2 := NewAtom("s", a)
	match, _ := MGU(atom1, atom2)
	if match {
		t.Error("expected match=false for different relations")
	}
}

// TestTermsMGUEqBasic tests MGU with equality remainder.
func TestTermsMGUEqBasic(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	x := mkVar("X")

	match, subs, eqs := TermsMGUEq(
		[]ResolutionTerm{x, a},
		[]ResolutionTerm{a, b},
	)
	if !match {
		t.Fatal("expected match=true")
	}
	if subs["X"].String() != "a" {
		t.Errorf("X: expected a, got %s", subs["X"])
	}
	if len(eqs) != 1 {
		t.Fatalf("expected 1 equality, got %d", len(eqs))
	}
	if eqs[0].RelName != "=" {
		t.Errorf("expected equality atom, got %s", eqs[0].RelName)
	}
}

// TestTermsMGUEqNoRemainder tests that matching constants produce no equalities.
func TestTermsMGUEqNoRemainder(t *testing.T) {
	a := mkConst("a")
	match, _, eqs := TermsMGUEq([]ResolutionTerm{a}, []ResolutionTerm{a})
	if !match {
		t.Fatal("expected match=true")
	}
	if len(eqs) != 0 {
		t.Errorf("expected 0 equalities, got %d", len(eqs))
	}
}

// TestMGUEqDifferentRelation tests MGUEq fails on different relation names.
func TestMGUEqDifferentRelation(t *testing.T) {
	a := mkConst("a")
	atom1 := NewAtom("r", a)
	atom2 := NewAtom("s", a)
	match, _, _ := MGUEq(atom1, atom2)
	if match {
		t.Error("expected match=false")
	}
}

// TestMGUEqMultipleEqualities tests collecting multiple equalities.
func TestMGUEqMultipleEqualities(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")
	d := mkConst("d")
	atom1 := NewAtom("r", a, c)
	atom2 := NewAtom("r", b, d)
	match, _, eqs := MGUEq(atom1, atom2)
	if !match {
		t.Fatal("expected match=true")
	}
	if len(eqs) != 2 {
		t.Fatalf("expected 2 equalities, got %d", len(eqs))
	}
}

// TestAtomFromApply tests extracting an Atom from an Apply node.
func TestAtomFromApply(t *testing.T) {
	rel := logic.NewConst("r", logic.TopS)
	a := mkConst("a")
	app, err := logic.NewApply(rel, a)
	if err != nil {
		t.Fatal(err)
	}
	atom := AtomFromApply(app)
	if atom == nil {
		t.Fatal("expected non-nil atom")
	}
	if atom.RelName != "r" {
		t.Errorf("expected relname r, got %s", atom.RelName)
	}
	if len(atom.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(atom.Args))
	}
}

// TestIsConstant tests the IsConstant predicate.
func TestIsConstant(t *testing.T) {
	c := mkConst("a")
	v := mkVar("X")
	if !ResolutionIsConstant(c) {
		t.Error("expected true for Const")
	}
	if ResolutionIsConstant(v) {
		t.Error("expected false for Var")
	}
}

// TestTermsMGUTransitiveChain tests a chain of variable bindings: X->Y->a.
func TestTermsMGUTransitiveChain(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	a := mkConst("a")
	match, subs := TermsMGU(
		[]ResolutionTerm{x, y},
		[]ResolutionTerm{y, a},
	)
	if !match {
		t.Fatal("expected match=true")
	}
	if subs["X"].String() != "a" {
		t.Errorf("X: expected a, got %s", subs["X"])
	}
	if subs["Y"].String() != "a" {
		t.Errorf("Y: expected a, got %s", subs["Y"])
	}
}

// FuzzTermsMGU fuzz-tests TermsMGU with random var/const selections.
func FuzzTermsMGU(f *testing.F) {
	// Seed corpus
	f.Add("X,a", "a,X")
	f.Add("X,Y", "Y,X")
	f.Add("a", "b")
	f.Add("a,b", "a,b")
	f.Add("X", "X")

	f.Fuzz(func(t *testing.T, s1, s2 string) {
		parts1 := strings.Split(s1, ",")
		parts2 := strings.Split(s2, ",")

		toTerm := func(s string) ResolutionTerm {
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				return mkConst("_empty")
			}
			if s[0] >= 'A' && s[0] <= 'Z' {
				v, err := logic.NewVariable(s, logic.TopS)
				if err != nil {
					return mkConst(s)
				}
				return v
			}
			return mkConst(s)
		}

		terms1 := make([]ResolutionTerm, len(parts1))
		terms2 := make([]ResolutionTerm, len(parts2))
		for i, p := range parts1 {
			terms1[i] = toTerm(p)
		}
		for i, p := range parts2 {
			terms2[i] = toTerm(p)
		}

		match, subs := TermsMGU(terms1, terms2)
		if match && subs != nil {
			// Verify substitution is consistent: each key maps to a resolved term
			for k, v := range subs {
				_ = k
				// The result should not be a variable that is also in the env
				if vr, ok := v.(*logic.Variable); ok {
					if _, found := subs[vr.Name]; found && vr.Name != k {
						// This would indicate an un-resolved chain
						t.Errorf("unresolved chain: %s -> %s -> ...", k, vr.Name)
					}
				}
			}
		}
	})
}
