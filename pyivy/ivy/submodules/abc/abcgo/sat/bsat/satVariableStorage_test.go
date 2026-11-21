package bsat

import (
	"slices"
	"testing"
	"unsafe"
)

func TestSATLoadFlagsKeepNativeWidthAndResetOnRestart(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_setnvars(s, 100)
	if width := unsafe.Sizeof(s.loads[0]); width != 1 {
		t.Errorf("load flags use %d bytes per variable, want 1", width)
	}
	s.loads[90] = 3
	s.tags[90] = 7
	s.polarity[90] = 1
	s.reasons[90] = 199
	s.model[90] = -1
	sat_solver_restart(s)
	sat_solver_setnvars(s, 7)
	sat_solver_setnvars(s, 94)
	if s.loads[90] != 0 || s.tags[90] != 0 || s.polarity[90] != 0 || s.reasons[90] != 0 || s.model[90] != 0 || s.assigns[90] != satUnassigned {
		t.Fatal("regrowth exposed stale variable metadata")
	}
}

func TestSATBulkGrowthPreservesDecisionHeapOrder(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		s := sat_solver_new()
		sat_solver_setnvars(s, 12)
		s.VarActType = mode
		// Simulate root assignments and an activity rescale. Some old
		// variables are absent from the heap and others have small scores.
		heap := []int32{8, 4, 3, 9, 1}
		activity := make([]uint64, 79)
		for i, v := range heap {
			activity[v] = uint64(6 - i)
		}
		copy(s.activity, activity)
		for i := range s.orderpos {
			s.orderpos[i] = -1
		}
		copy(s.order.ptr, heap)
		s.order.size = int32(len(heap))
		for i, v := range heap {
			s.orderpos[v] = int32(i)
		}
		for v := int32(12); v < 79; v++ {
			if mode == 0 {
				activity[v] = 1 << 10
			}
			heap = append(heap, v)
			i := len(heap) - 1
			for i > 0 && activity[heap[(i-1)/2]] < activity[v] {
				heap[i] = heap[(i-1)/2]
				i = (i - 1) / 2
			}
			heap[i] = v
		}
		sat_solver_setnvars(s, 79)
		if !slices.Equal(veci_begin(&s.order), heap) || !slices.Equal(s.activity, activity) {
			t.Fatalf("activity mode %d changed heap order", mode)
		}
		for i, v := range heap {
			if s.orderpos[v] != int32(i) {
				t.Fatalf("mode %d variable %d has position %d, want %d", mode, v, s.orderpos[v], i)
			}
		}
		sat_solver_delete(s)
	}
}

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
