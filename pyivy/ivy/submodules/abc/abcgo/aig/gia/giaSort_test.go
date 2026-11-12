package gia

import "testing"

func TestGiaSortFloats(t *testing.T) {
	values := []float32{3, -2, 1, 3}
	original := append([]float32(nil), values...)
	perm := Gia_SortFloats(values, nil, int32(len(values)))
	for i := range values {
		if i > 0 && values[i-1] > values[i] {
			t.Fatal(values)
		}
		if original[perm[i]] != values[i] {
			t.Fatalf("invalid permutation %v", perm)
		}
	}
	values = []float32{2, 1, 3}
	perm = []int32{10, 11, 12}
	Gia_SortFloats(values, perm, 2)
	if values[0] != 1 || perm[0] != 11 || values[2] != 3 {
		t.Fatalf("partial sort %v %v", values, perm)
	}
}
