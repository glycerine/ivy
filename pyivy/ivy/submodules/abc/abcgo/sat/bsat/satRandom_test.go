package bsat

import "testing"

func TestSolverRandomDeterministic(t *testing.T) {
	a, b := 91648253.0, 91648253.0
	for i := 0; i < 10000; i++ {
		x, y := drand(&a), drand(&b)
		if x != y || x < 0 || x >= 1 {
			t.Fatal(i, x, y)
		}
	}
	a = 91648253
	for i := 0; i < 10000; i++ {
		x := irand(&a, 17)
		if x < 0 || x >= 17 {
			t.Fatal(x)
		}
	}
}
