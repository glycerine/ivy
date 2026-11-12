package glucose

import "testing"

func TestLiteralAndLBoolEncoding(t *testing.T) {
	for v := Var(0); v < 100; v++ {
		for _, sign := range []bool{false, true} {
			l := MkLit(v, sign)
			if l.Var() != v || l.Sign() != sign || l.Not().Not() != l {
				t.Fatal("literal encoding")
			}
		}
	}
	for a := LBool(0); a < 4; a++ {
		for b := LBool(0); b < 4; b++ {
			if a.And(b).IsUndef() {
				if a.Equal(LFalse) || b.Equal(LFalse) {
					t.Fatal("AND undef")
				}
			} else if a.And(b).Equal(LTrue) != (a.Equal(LTrue) && b.Equal(LTrue)) {
				t.Fatal("AND")
			}
			if a.Or(b).IsUndef() {
				if a.Equal(LTrue) || b.Equal(LTrue) {
					t.Fatal("OR undef")
				}
			} else if a.Or(b).Equal(LFalse) != (a.Equal(LFalse) && b.Equal(LFalse)) {
				t.Fatal("OR")
			}
		}
	}
	if !LUndef.Equal(LUndef.Not()) {
		t.Fatal("complemented undefined")
	}
}
func TestClauseSubsumptionAndStrengthen(t *testing.T) {
	a, b, c := MkLit(0, false), MkLit(1, false), MkLit(2, false)
	small := NewClause([]Lit{a, b}, false, true)
	large := NewClause([]Lit{a, b, c}, false, true)
	if small.Subsumes(large) != LitUndef {
		t.Fatal("full subsumption")
	}
	opposite := NewClause([]Lit{a.Not(), b, c}, false, true)
	if small.Subsumes(opposite) != a {
		t.Fatal("self-subsuming resolution")
	}
	opposite.Strengthen(a.Not())
	if len(opposite.Lits) != 2 || opposite.Abstraction() != (1<<1|1<<2) {
		t.Fatal("strengthen")
	}
	if small.Subsumes(opposite) != LitError {
		t.Fatal("missing variable")
	}
	both := NewClause([]Lit{a.Not(), b.Not(), b, c}, false, true)
	if small.Subsumes(both) != a {
		t.Fatal("must continue looking for direct literal")
	}
}
func TestAllocatorRelocationPreservesMetadata(t *testing.T) {
	var a, b ClauseAllocator
	a.ExtraClauseField = true
	c := a.Alloc([]Lit{MkLit(0, false), MkLit(1, true)}, true)
	c.SetActivity(12)
	c.SetLBD(2)
	c.SetCanBeDel(false)
	c.SetSizeWithoutSelectors(1)
	old := c
	a.Reloc(&c, &b)
	if c == old || old.Relocation() != c || c.Activity() != 12 || c.LBD() != 2 || c.CanBeDel() || c.SizeWithoutSelectors() != 1 {
		t.Fatal("metadata lost")
	}
	again := old
	a.Reloc(&again, &b)
	if again != c {
		t.Fatal("duplicate relocation")
	}
	before := b.Size()
	c.SetMark(1)
	b.Free(c)
	b.Free(c)
	if b.Wasted() != before {
		t.Fatal("free accounting")
	}
}
