package congclos

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

func mkConst(name string) *logic.Const {
	return logic.NewConst(name, logic.TopS)
}

// TestNewCongClos tests that a fresh CongClos is empty.
func TestNewCongClos(t *testing.T) {
	cc := New()
	if cc == nil {
		t.Fatal("expected non-nil CongClos")
	}
	theory := cc.Theory()
	if len(theory) != 0 {
		t.Errorf("expected empty theory, got %d entries", len(theory))
	}
}

// TestFindCreatesEntry tests that Find creates a node if not present.
func TestFindCreatesEntry(t *testing.T) {
	cc := New()
	v := mkConst("v")
	rep := cc.Find(v)
	if rep.String() != "v" {
		t.Errorf("expected v, got %s", rep)
	}
}

// TestFindByName tests looking up by name.
func TestFindByName(t *testing.T) {
	cc := New()
	rep := cc.FindByName("v")
	if rep.String() != "v" {
		t.Errorf("expected v, got %s", rep)
	}
}

// TestUnionBasic tests basic union operation (from Python docstring).
func TestUnionBasic(t *testing.T) {
	cc := New()
	v := mkConst("v")
	w := mkConst("w")
	cc.Union(v, w)
	rep := cc.Find(w)
	if rep.String() != "v" {
		t.Errorf("expected v as representative, got %s", rep)
	}
}

// TestUnionMinimalRep tests that the lexicographically smaller name is representative.
func TestUnionMinimalRep(t *testing.T) {
	cc := New()
	a := mkConst("a")
	z := mkConst("z")
	cc.Union(z, a)
	if cc.Find(z).String() != "a" {
		t.Errorf("expected a as rep of z, got %s", cc.Find(z))
	}
	if cc.Find(a).String() != "a" {
		t.Errorf("expected a as rep of a, got %s", cc.Find(a))
	}
}

// TestUnionTransitive tests transitivity: union(a,b), union(b,c) => find(c)==a.
func TestUnionTransitive(t *testing.T) {
	cc := New()
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")
	cc.Union(a, b)
	cc.Union(b, c)
	if cc.Find(c).String() != "a" {
		t.Errorf("expected a as rep of c, got %s", cc.Find(c))
	}
}

// TestUnionSameClass tests that union of already-unified terms is a no-op.
func TestUnionSameClass(t *testing.T) {
	cc := New()
	a := mkConst("a")
	b := mkConst("b")
	cc.Union(a, b)
	cc.Union(a, b) // second time, should be no-op
	if cc.Find(b).String() != "a" {
		t.Errorf("expected a, got %s", cc.Find(b))
	}
}

// TestTheory tests that Theory reports equalities.
func TestTheory(t *testing.T) {
	cc := New()
	v := mkConst("v")
	w := mkConst("w")
	cc.Union(v, w)
	theory := cc.Theory()
	if len(theory) != 1 {
		t.Fatalf("expected 1 equality, got %d", len(theory))
	}
	lit := theory[0]
	if lit.Polarity != 1 {
		t.Error("expected positive polarity")
	}
	if lit.Atom.RelName != "=" {
		t.Errorf("expected =, got %s", lit.Atom.RelName)
	}
}

// TestPushPop tests that push/pop correctly restores state.
func TestPushPop(t *testing.T) {
	cc := New()
	v := mkConst("v")
	w := mkConst("w")
	u := mkConst("u")

	cc.Union(v, w)
	cc.Push()
	cc.Union(u, v)

	// Before pop: u, v, w should all have rep "u"
	if cc.Find(w).String() != "u" {
		t.Errorf("before pop: expected u as rep of w, got %s", cc.Find(w))
	}

	cc.Pop()

	// After pop: u should be independent, v and w should still be unified
	if cc.Find(w).String() != "v" {
		t.Errorf("after pop: expected v as rep of w, got %s", cc.Find(w))
	}
	if cc.Find(u).String() != "u" {
		t.Errorf("after pop: expected u as rep of u, got %s", cc.Find(u))
	}
}

// TestMultiplePushPop tests nested push/pop.
func TestMultiplePushPop(t *testing.T) {
	cc := New()
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")

	cc.Push()
	cc.Union(a, b)
	cc.Push()
	cc.Union(b, c)

	if cc.Find(c).String() != "a" {
		t.Errorf("expected a, got %s", cc.Find(c))
	}

	cc.Pop() // undo union(b,c)
	if cc.Find(c).String() != "c" {
		t.Errorf("after first pop: expected c, got %s", cc.Find(c))
	}
	if cc.Find(b).String() != "a" {
		t.Errorf("after first pop: expected a as rep of b, got %s", cc.Find(b))
	}

	cc.Pop() // undo union(a,b)
	if cc.Find(b).String() != "b" {
		t.Errorf("after second pop: expected b, got %s", cc.Find(b))
	}
}

// TestPopEmpty tests that Pop on empty pushes is safe.
func TestPopEmpty(t *testing.T) {
	cc := New()
	cc.Pop() // should not panic
}

// TestFindByNameCreates tests that FindByName creates a constant.
func TestFindByNameCreates(t *testing.T) {
	cc := New()
	rep := cc.FindByName("newconst")
	if rep.String() != "newconst" {
		t.Errorf("expected newconst, got %s", rep)
	}
}

// TestFindByNameExisting tests FindByName with an existing entry.
func TestFindByNameExisting(t *testing.T) {
	cc := New()
	a := mkConst("a")
	b := mkConst("b")
	cc.Union(a, b)
	rep := cc.FindByName("b")
	if rep.String() != "a" {
		t.Errorf("expected a, got %s", rep)
	}
}

// TestTheoryEmpty tests that an empty CongClos has empty theory.
func TestTheoryEmpty(t *testing.T) {
	cc := New()
	theory := cc.Theory()
	if len(theory) != 0 {
		t.Errorf("expected 0, got %d", len(theory))
	}
}

// TestUnionManyElements tests union with many elements.
func TestUnionManyElements(t *testing.T) {
	cc := New()
	consts := make([]*logic.Const, 10)
	for i := range consts {
		consts[i] = mkConst(fmt.Sprintf("c%d", i))
	}
	for i := 1; i < len(consts); i++ {
		cc.Union(consts[0], consts[i])
	}
	for _, c := range consts {
		rep := cc.Find(c)
		if rep.String() != "c0" {
			t.Errorf("expected c0 as rep, got %s for %s", rep, c)
		}
	}
}

// TestPushPopPreservesTab tests that tab entries persist after pop (only reps change).
func TestPushPopPreservesTab(t *testing.T) {
	cc := New()
	a := mkConst("a")
	cc.Push()
	cc.Find(a) // creates entry in tab
	cc.Pop()
	// The entry should still be in tab
	rep := cc.Find(a)
	if rep.String() != "a" {
		t.Errorf("expected a, got %s", rep)
	}
}

// FuzzCongClos fuzz-tests union-find consistency.
func FuzzCongClos(f *testing.F) {
	f.Add("a,b;b,c;?c", "")
	f.Add("x,y;y,z;push;z,w;pop;?w", "")

	f.Fuzz(func(t *testing.T, ops, _ string) {
		cc := New()
		pushCount := 0
		for _, op := range strings.Split(ops, ";") {
			op = strings.TrimSpace(op)
			if op == "" {
				continue
			}
			if op == "push" {
				cc.Push()
				pushCount++
				continue
			}
			if op == "pop" {
				if pushCount > 0 {
					cc.Pop()
					pushCount--
				}
				continue
			}
			if strings.HasPrefix(op, "?") {
				name := strings.TrimPrefix(op, "?")
				if len(name) > 0 && len(name) < 20 {
					cc.FindByName(name)
				}
				continue
			}
			parts := strings.SplitN(op, ",", 2)
			if len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 &&
				len(parts[0]) < 20 && len(parts[1]) < 20 {
				cc.Union(mkConst(parts[0]), mkConst(parts[1]))
			}
		}
		// Pop remaining
		for pushCount > 0 {
			cc.Pop()
			pushCount--
		}
	})
}
