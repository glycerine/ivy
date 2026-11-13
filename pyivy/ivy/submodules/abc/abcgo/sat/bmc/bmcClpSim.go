package bmc

import (
	"fmt"
	"math/rand"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_ComputeSimDiff returns output zero's value and marks inputs whose
// individual inversion changes it.
func Bmc_ComputeSimDiff(p *gia.Gia_Man_t, pattern, differences *vec.Vec_Int_t) int32 {
	if p == nil || pattern == nil || differences == nil || vec.Vec_IntSize(pattern) != int(gia.Gia_ManCiNum(p)) || gia.Gia_ManCoNum(p) == 0 {
		panic("bmc: invalid simulation pattern")
	}
	input := make([]bool, vec.Vec_IntSize(pattern))
	for i, v := range vec.Vec_IntArray(pattern) {
		if v != 0 && v != 1 {
			panic("bmc: nonbinary simulation pattern")
		}
		input[i] = v != 0
	}
	base, err := p.Evaluate(input)
	if err != nil {
		panic(err)
	}
	vec.Vec_IntClear(differences)
	for i := range input {
		input[i] = !input[i]
		out, err := p.Evaluate(input)
		if err != nil {
			panic(err)
		}
		var diff int32
		if out[0] != base[0] {
			diff = 1
		}
		vec.Vec_IntPush(differences, diff)
		input[i] = !input[i]
	}
	if base[0] {
		return 1
	}
	return 0
}
func Bmc_ComputeSimTest(p *gia.Gia_Man_t) {
	if p == nil {
		return
	}
	random := rand.New(rand.NewSource(1))
	pattern, diffs := vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(p))), vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(p)))
	counts := make([]int, int(gia.Gia_ManCiNum(p)))
	for i := 0; i < 256; i++ {
		vec.Vec_IntClear(pattern)
		for range counts {
			vec.Vec_IntPush(pattern, int32(random.Intn(2)))
		}
		Bmc_ComputeSimDiff(p, pattern, diffs)
		for j, x := range vec.Vec_IntArray(diffs) {
			counts[j] += int(x)
		}
	}
	fmt.Printf("Number of patterns = 256.\nInput differences: %v\n", counts)
}
