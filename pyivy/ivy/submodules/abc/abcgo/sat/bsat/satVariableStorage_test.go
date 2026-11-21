package bsat

import (
	"testing"
	"unsafe"
)

func TestSATVariableFlagsUseNativeBytes(t *testing.T) {
	a := newSatAssignments(1)
	b := unsafe.Sizeof(a.assigns[0]) + unsafe.Sizeof(a.polarity[0]) + unsafe.Sizeof(a.tags[0])
	if b != 3 {
		t.Fatalf("assignment, polarity and tag consume %d bytes per variable, want 3", b)
	}
	if a.assigns[0] != satUnassigned || a.polarity[0] != 0 || a.tags[0] != 0 {
		t.Fatal("initial flags")
	}
}

func TestSATGrowPreservesFlagsAndFullWidthReasons(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_setnvars(s, 5)
	if !s.enqueue(toLitCond(3, 1), 1<<26) {
		t.Fatal("enqueue")
	}
	s.polarity[4] = 1
	s.tagVariable(3, 15)
	sat_solver_setnvars(s, 1025)
	if s.assigns[3] != 1 || s.reasons[3] != 1<<26 || s.tags[3] != 15 || s.polarity[4] != 1 || s.assigns[1024] != satUnassigned {
		t.Fatal("lost flags or clause handle during growth")
	}
	s.clearTags(0)
	if s.tags[3] != 0 || s.polarity[4] != 1 || s.assigns[3] != 1 {
		t.Fatal("tag reset changed another flag")
	}
}

func TestSATEnqueuePreservesNonzeroFrequencies(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_setnvars(s, 3)
	s.pFreqs[0] = 23
	if sat_solver_enqueue(s, toLit(0), 0) != 1 || s.pFreqs[0] != 23 {
		t.Errorf("direct enqueue frequency %d, want 23", s.pFreqs[0])
	}
	s.pFreqs[0] = 255
	if sat_solver_enqueue(s, lit_neg(toLit(0)), 0) != 0 || s.pFreqs[0] != 255 {
		t.Errorf("conflicting enqueue frequency %d, want 255", s.pFreqs[0])
	}
	veci_push(&s.wlists[toLit(0)], clause_from_lit(toLit(1)))
	veci_push(&s.wlists[toLit(1)], clause_from_lit(toLit(2)))
	s.pFreqs[1] = 17
	if sat_solver_propagate(s) != 0 || s.pFreqs[1] != 17 || s.pFreqs[2] != 1 {
		t.Errorf("implied frequencies %v, want [17 1]", s.pFreqs[1:3])
	}
}

func BenchmarkSATVariableSetup65536(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := sat_solver_new()
		sat_solver_setnvars(s, 65536)
		if s.assigns[65535] != satUnassigned || s.tags[65535] != 0 || s.polarity[65535] != 0 {
			b.Fatal("initial variable state")
		}
		sat_solver_delete(s)
	}
}
