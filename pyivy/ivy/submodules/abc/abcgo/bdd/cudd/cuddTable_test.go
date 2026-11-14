package cudd

import "testing"

func TestCuddPrime(t *testing.T) {
	tests := []struct {
		input, want uint32
	}{
		{0, 2},
		{1, 2},
		{2, 2},
		{3, 3},
		{4, 5},
		{17, 17},
		{18, 19},
		{24, 29},
		{100, 101},
		{25, 29},
		{49, 53},
		{121, 127},
		{4294967291, 4294967291}, // largest prime representable by uint32
		{4294967295, 0},          // no representable prime at or above this value
	}
	for _, tt := range tests {
		if got := Cudd_Prime(tt.input); got != tt.want {
			t.Errorf("Cudd_Prime(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCuddPrimeUint32Boundaries(t *testing.T) {
	for _, input := range []uint32{0, 1, 2} {
		if got := Cudd_Prime(input); got != 2 {
			t.Errorf("Cudd_Prime(%d) = %d, want 2", input, got)
		}
	}
}
