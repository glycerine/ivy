package gia

import "testing"

func TestRsb2WindowOptimization(t *testing.T) {
	// (a & b) & a contains a redundant second AND.
	objects := []int32{0, 0, 0, 0, 0, 0, 2, 4, 2, 6, 8, 8}
	updated, changes := Abc_ResubComputeWindow(objects, 6, 20, -1, 1, 0, 1, 0)
	if changes == 0 || len(updated) >= len(objects) {
		t.Fatalf("no improvement: %v, %d", updated, changes)
	}
	before := Gia_ManFromResub(objects, 6, 2)
	after := Gia_ManFromResub(updated, int32(len(updated)/2), 2)
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		a, e := before.Evaluate(input)
		b, f := after.Evaluate(input)
		if e != nil || f != nil || a[0] != b[0] {
			t.Fatalf("%v -> %v vs %v", input, a, b)
		}
	}
}
