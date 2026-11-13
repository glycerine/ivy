package mpm

import (
	"math/rand"
	"testing"
)

func TestVariablePermutationInverses(t *testing.T) {
	v, p := make([]int32, 6), make([]int32, 6)
	Mpm_VarsClear(v, p, 6)
	rng := rand.New(rand.NewSource(3461))
	for step := 0; step < 1000; step++ {
		a := int32(rng.Intn(6))
		b := (a + 1 + int32(rng.Intn(5))) % 6
		Mpm_VarsSwap(v, p, a, b)
		for i := int32(0); i < 6; i++ {
			if v[p[i]] != i || p[v[i]] != i {
				t.Fatalf("step %d broken inverses %v %v", step, v, p)
			}
		}
	}
}
