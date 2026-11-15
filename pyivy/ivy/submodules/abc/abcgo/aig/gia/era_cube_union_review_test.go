package gia

import (
	"fmt"
	"slices"
	"testing"
)

func cubeReviewStorage(t *testing.T, registers int32) *Gia_ManAre_t {
	t.Helper()
	g := Gia_ManStart(2*registers + 1)
	var r []int32
	for i := int32(0); i < registers; i++ {
		r = append(r, Gia_ManAppendCi(g))
	}
	for _, lit := range r {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, registers)
	return Gia_ManAreCreate(g)
}

// Inspect rooted storage independently of the cube relation/counting helpers.
// A Boolean truth set is the oracle; list/tree shape and IDs may differ.
func cubeReviewUnion(t *testing.T, p *Gia_ManAre_t, positions []int32) []bool {
	t.Helper()
	got := make([]bool, 1<<len(positions))
	seenNodes, seenStates := map[uint32]bool{}, map[uint32]bool{}
	var walk func(Gia_PtrAre_t, bool)
	walk = func(handle Gia_PtrAre_t, tree bool) {
		id := uint32(handle) & 0x7fffffff
		if tree {
			if id == 0 || int(id) >= len(p.objects) || seenNodes[id] {
				t.Fatalf("invalid/cyclic tree handle%d", id)
			}
			seenNodes[id] = true
			node := p.objects[id]
			counts := [3]uint8{node.nStas0, node.nStas1, node.nStas2}
			for i, child := range node.F {
				walk(child, counts[i] == 0 && uint32(child)&0x7fffffff != 0)
			}
			return
		}
		for id != 0 {
			if int(id) >= len(p.states) || seenStates[id] {
				t.Fatalf("invalid/duplicate list handle%d", id)
			}
			seenStates[id] = true
			state := p.states[id]
			if uint32(state.iPrev)>>31 == 0 {
				for assignment := range got {
					matches := true
					for i, pos := range positions {
						trit := (state.pData[(2*pos)>>5] >> uint((2*pos)&31)) & 3
						if trit == 3 {
							t.Fatal("contradictory cube bits")
						}
						if trit == 1 && assignment&(1<<i) != 0 || trit == 2 && assignment&(1<<i) == 0 {
							matches = false
							break
						}
					}
					got[assignment] = got[assignment] || matches
				}
			}
			id = uint32(state.iNext) & 0x7fffffff
		}
	}
	walk(p.Root, p.fTree != 0)
	for id, state := range p.states[1:] {
		if uint32(state.iPrev)>>31 == 0 && !seenStates[uint32(id+1)] {
			t.Fatalf("live state%d lost from root", id+1)
		}
	}
	return got
}

func cubeReviewAddCode(p *Gia_ManAre_t, code int, positions []int32) *Gia_StaAre_t {
	state := Gia_ManAreCreateSta(p)
	for _, pos := range positions {
		switch code % 3 {
		case 1:
			Gia_StaSetValue0(state, pos)
		case 2:
			Gia_StaSetValue1(state, pos)
		}
		code /= 3
	}
	return state
}
func cubeReviewAddTruth(want []bool, code int, positions []int32) {
	for assignment := range want {
		digits, matches := code, true
		for i := range positions {
			trit := digits % 3
			digits /= 3
			if trit == 1 && assignment&(1<<i) != 0 || trit == 2 && assignment&(1<<i) == 0 {
				matches = false
				break
			}
		}
		want[assignment] = want[assignment] || matches
	}
}

func TestCubeReachabilityExhaustiveThreeCubeUnions(t *testing.T) {
	p := cubeReviewStorage(t, 3)
	defer Gia_ManAreFree(p)
	positions := []int32{0, 1, 2}
	// All 27^3 ordered cube triples, including containment, disjointness,
	// sharping after earlier comparisons, duplicate recycling and unused rows.
	for a := 0; a < 27; a++ {
		for b := 0; b < 27; b++ {
			for c := 0; c < 27; c++ {
				p.states, p.objects = nil, nil
				p.nStas, p.nObjs, p.nStaPages, p.nObjPages = 0, 0, 0, 0
				p.Root, p.fTree = 0, 0
				want := make([]bool, 8)
				for step, code := range []int{a, b, c} {
					cubeReviewAddTruth(want, code, positions)
					Gia_ManAreCubeProcess(p, cubeReviewAddCode(p, code, positions))
					if got := cubeReviewUnion(t, p, positions); !slices.Equal(got, want) {
						t.Fatalf("triple%d/%d/%d step%d union%v want%v", a, b, c, step, got, want)
					}
				}
			}
		}
	}
}

func TestCubeReachabilityMixedTreeUnions(t *testing.T) {
	for seed := uint32(1); seed <= 16; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			p := cubeReviewStorage(t, 8)
			defer Gia_ManAreFree(p)
			positions := []int32{0, 1, 2, 3, 4, 5, 6, 7}
			want := make([]bool, 256)
			// Seed disjoint minterms with holes so mixed cubes still contribute
			// new states after the initial and nested tree split thresholds.
			for mask := 0; mask < 256; mask++ {
				if mask%4 == 3 {
					continue
				}
				code, mult := 0, 1
				for i := 0; i < 8; i++ {
					code += (1 + ((mask >> i) & 1)) * mult
					mult *= 3
				}
				cubeReviewAddTruth(want, code, positions)
				Gia_ManAreCubeProcess(p, cubeReviewAddCode(p, code, positions))
			}
			if p.fTree == 0 {
				t.Fatal("tree threshold was not exercised")
			}
			rng := seed
			for step := 0; step < 100; step++ {
				rng = rng*1664525 + 1013904223
				code := int(rng % 6561)
				cubeReviewAddTruth(want, code, positions)
				p.iStaCur = int32(step) % p.nStas // exercise processed/unprocessed collections
				Gia_ManAreCubeProcess(p, cubeReviewAddCode(p, code, positions))
				if got := cubeReviewUnion(t, p, positions); !slices.Equal(got, want) {
					t.Fatalf("step%d cube%d union mismatch", step, code)
				}
			}
		})
	}
}
