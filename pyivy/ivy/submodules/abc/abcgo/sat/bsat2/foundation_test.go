package bsat2

import (
	"math/rand"
	"testing"
)

func TestVectorOwnershipAndGrowth(t *testing.T) {
	var v Vec[*int]
	for i := 0; i < 2000; i++ {
		value := i
		v.Push(&value)
	}
	var copy Vec[*int]
	v.CopyTo(&copy)
	if copy.Len() != 2000 || copy.At(1999) != v.At(1999) {
		t.Fatal("copy")
	}
	v.Shrink(1999)
	if v.data[:cap(v.data)][1] != nil || copy.Len() != 2000 {
		t.Fatal("ownership or retained reference")
	}
	copy.MoveTo(&v)
	if copy.Len() != 0 || v.Len() != 2000 {
		t.Fatal("move")
	}
	for i := 1999; i >= 0; i-- {
		if *v.Pop() != i {
			t.Fatal(i)
		}
	}
}
func TestRegionReferencesSurviveGrowthAndMove(t *testing.T) {
	r := NewRegionAllocator[uint32](2)
	first := r.Alloc(2)
	r.Slice(first, 2)[0] = 123
	for i := 0; i < 1000; i++ {
		r.Alloc(3)
	}
	if *r.At(first) != 123 || r.Size() != 3002 {
		t.Fatal("stable offset")
	}
	r.Free(2)
	var moved RegionAllocator[uint32]
	r.MoveTo(&moved)
	if r.Size() != 0 || moved.Size() != 3002 || moved.Wasted() != 2 || *moved.At(first) != 123 {
		t.Fatal("region move")
	}
}
func TestIndexedHeapMutablePriorities(t *testing.T) {
	rng := rand.New(rand.NewSource(738))
	priorities := make([]int, 100)
	less := func(a, b int32) bool {
		if priorities[a] == priorities[b] {
			return a < b
		}
		return priorities[a] < priorities[b]
	}
	h := NewHeap(less)
	for i := range priorities {
		priorities[i] = rng.Intn(10000)
		h.Insert(int32(i))
	}
	for i := 0; i < 1000; i++ {
		v := rng.Intn(100)
		priorities[v] = rng.Intn(10000)
		h.Update(int32(v))
	}
	previous := int32(-1)
	for h.Len() > 0 {
		v := h.RemoveMin()
		if previous >= 0 && less(v, previous) {
			t.Fatal("heap ordering")
		}
		if h.Contains(v) {
			t.Fatal("stale membership")
		}
		previous = v
	}
	h.Build([]int32{3, 8, 1})
	h.Clear(false)
	if h.Len() != 0 || h.Contains(3) {
		t.Fatal("clear")
	}
}

func TestClauseSubsumptionAndRelocation(t *testing.T) {
	a := NewClauseAllocator()
	a.ExtraClauseField = true
	first := a.Alloc([]Lit{0, 2}, false)
	other := a.Alloc([]Lit{1, 2, 4}, false)
	if a.At(first).Subsumes(a.At(other)) != 0 {
		t.Fatal("self-subsuming resolution")
	}
	a.At(other).Strengthen(1)
	if a.At(first).Subsumes(a.At(other)) != LitError {
		t.Fatal("abstraction updated")
	}
	dest := NewClauseAllocator()
	dest.ExtraClauseField = true
	old := first
	a.Relocate(&first, dest)
	again := old
	a.Relocate(&again, dest)
	if first != again || len(dest.At(first).Lits) != 2 || dest.Size() != 4 {
		t.Fatal("relocation identity")
	}
}
func TestLiftedBooleanTruthTables(t *testing.T) {
	for _, a := range []LBool{True, False, Undef} {
		for _, b := range []LBool{True, False, Undef} {
			and := Undef
			if a == False || b == False {
				and = False
			} else if a == True && b == True {
				and = True
			}
			or := Undef
			if a == True || b == True {
				or = True
			} else if a == False && b == False {
				or = False
			}
			if a.And(b) != and || a.Or(b) != or {
				t.Fatal(a, b, a.And(b), a.Or(b))
			}
		}
	}
}

func TestCoreClausePropagationAndBacktrack(t *testing.T) {
	s := NewSolver()
	s.AddClause(1, 2)
	s.AddClause(1, 3)
	s.trailLimits = append(s.trailLimits, len(s.trail))
	s.uncheckedEnqueue(0, CRefUndef)
	if conflict := s.propagate(); conflict == CRefUndef {
		t.Fatal("missing conflict")
	}
	s.cancelUntil(0)
	if s.Value(0) != Undef || s.Value(2) != Undef || s.level() != 0 {
		t.Fatal("backtrack")
	}
	if !s.AddClause(1) || s.Value(1) != True {
		t.Fatal("root unit")
	}
}
