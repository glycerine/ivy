package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"testing"
)

func TestCutTruthRandomFunctionsAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(43211))
	for trial := 0; trial < 500; trial++ {
		p := &Sdb_Sto_t{nCutSize: 6, fTruthMin: int32(trial % 2), vTtMem: vec.Vec_MemAllocForTT(6, 0)}
		ma, mb := rng.Intn(64), rng.Intn(64)
		a, b, out := testCut(ma), testCut(mb), testCut(ma|mb)
		// Every truth-table word repeats its function at the cut's arity.
		makeFunc := func(c *Sdb_Cut_t) uint64 {
			bits := rng.Uint64()
			var truth uint64
			for i := 0; i < 64; i++ {
				idx := i & ((1 << c.nLeaves) - 1)
				if bits&(uint64(1)<<idx) != 0 {
					truth |= uint64(1) << i
				}
			}
			phase := int32(truth & 1)
			if phase != 0 {
				truth = ^truth
			}
			id := vec.Vec_MemHashInsert(p.vTtMem, []uint64{truth})
			c.iFunc = int32(id)*2 + phase
			return truth ^ (uint64(0) - uint64(phase))
		}
		ta, tb := makeFunc(a), makeFunc(b)
		ca, cb, xor := int32(rng.Intn(2)), int32(rng.Intn(2)), int32(rng.Intn(2))
		eval := func(c *Sdb_Cut_t, truth uint64, assignment int) bool {
			index := 0
			for i := uint32(0); i < c.nLeaves; i++ {
				if assignment&(1<<c.pLeaves[i]) != 0 {
					index |= 1 << i
				}
			}
			return truth&(uint64(1)<<index) != 0
		}
		want := func(assignment int) bool {
			x := eval(a, ta, assignment) != (ca != 0)
			y := eval(b, tb, assignment) != (cb != 0)
			if xor != 0 {
				return x != y
			}
			return x && y
		}
		old := out.nLeaves
		shrunk := Sdb_CutComputeTruth(p, a, b, ca, cb, out, xor)
		truth := Sdb_CutTruth(p, out)
		if out.iFunc&1 != 0 {
			truth = ^truth
		}
		for assignment := 0; assignment < 64; assignment++ {
			if got := eval(out, truth, assignment); got != want(assignment) {
				t.Fatalf("trial %d assignment %d masks %x/%x got %t", trial, assignment, ma, mb, got)
			}
		}
		if (shrunk != 0) != (out.nLeaves < old) {
			t.Fatalf("trial %d incorrect shrink flag", trial)
		}
		if p.fTruthMin != 0 {
			for i := 0; i < 6; i++ {
				needed := false
				for assignment := 0; assignment < 64; assignment++ {
					if want(assignment) != want(assignment^(1<<i)) {
						needed = true
						break
					}
				}
				found := false
				for j := uint32(0); j < out.nLeaves; j++ {
					found = found || out.pLeaves[j] == int32(i)
				}
				if needed != found {
					t.Fatalf("trial %d variable %d needed %t retained %t", trial, i, needed, found)
				}
			}
		}
	}
}
