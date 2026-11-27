package ioa

import "testing"

// A model with several cycles must always report the same one (gates are
// visited in variable order), not whichever Go's map iteration finds first.
func TestCycleErrorIsDeterministic(t *testing.T) {
	src := []byte("aag 9 1 0 1 8\n2\n18\n4 6 2\n6 4 2\n8 10 2\n10 8 2\n12 14 2\n14 12 2\n16 18 2\n18 16 2\n")
	var want string
	for run := 0; run < 20; run++ {
		_, err := Read(src)
		if err == nil {
			t.Fatal("cyclic model accepted")
		}
		if run == 0 {
			want = err.Error()
		} else if err.Error() != want {
			t.Fatalf("run %d: %q, want %q", run, err, want)
		}
	}
}

// Symbol checks follow AIGER section order, so with several bad symbols
// the first in that order is reported.
func TestSymbolCheckFollowsSectionOrder(t *testing.T) {
	m := &Model{MaxVar: 1, Inputs: []uint32{2}, Symbols: map[Symbol]string{}}
	for _, s := range []Symbol{{'x', 0}, {'o', 0}, {'l', 3}, {'i', 5}} {
		m.Symbols[s] = "n"
	}
	for run := 0; run < 20; run++ {
		err := m.Check()
		if err == nil || err.Error() != "symbol index out of range" {
			t.Fatalf("run %d: got %v, want the i5 index error first", run, err)
		}
	}
}
