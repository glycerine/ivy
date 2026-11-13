package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"math/rand"
	"testing"
)

func TestCec2SweepPreservesOutputs(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for trial := 0; trial < 30; trial++ {
		for _, cones := range []bool{false, true} {
			g := gia.Gia_ManStart(64)
			lits := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
			for i := 0; i < 14; i++ {
				ai := r.Intn(len(lits))
				bi := (ai + 1 + r.Intn(len(lits)-1)) % len(lits)
				a, b := lits[ai], lits[bi]
				a ^= int32(r.Intn(2))
				b ^= int32(r.Intn(2))
				lits = append(lits, gia.Gia_ManAppendAnd(g, a, b))
			}
			for i := 0; i < 5; i++ {
				gia.Gia_ManAppendCo(g, lits[4+r.Intn(14)]^int32(r.Intn(2)))
			}
			reduced, _, _ := cec2SweepEquivalences(g, cec2SweepSettings{finishGraph: true, words: 2, rounds: 1, iterations: 5, conflictLimit: 1000, useCones: cones})
			for mask := 0; mask < 16; mask++ {
				in := make([]bool, 4)
				for i := range in {
					in[i] = mask&(1<<i) != 0
				}
				a, e := g.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				b, e := reduced.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				for i := range a {
					if a[i] != b[i] {
						t.Fatalf("trial=%d cones=%v input=%04b output=%d got=%v want=%v", trial, cones, mask, i, b[i], a[i])
					}
				}
			}
		}
	}
}

func TestCec2SweepMergesAssociativeAnds(t *testing.T) {
	for _, cones := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		xy := gia.Gia_ManAppendAnd(g, x, y)
		yz := gia.Gia_ManAppendAnd(g, y, z)
		a := gia.Gia_ManAppendAnd(g, xy, z)
		b := gia.Gia_ManAppendAnd(g, x, yz)
		gia.Gia_ManAppendCo(g, a)
		gia.Gia_ManAppendCo(g, b)
		out, _, _ := cec2SweepEquivalences(g, cec2SweepSettings{finishGraph: true, words: 2, rounds: 1, iterations: 4, conflictLimit: 1000, useCones: cones})
		left, _, _, _ := out.Gate(gia.Gia_ObjId(out, gia.Gia_ManCo(out, 0)))
		right, _, _, _ := out.Gate(gia.Gia_ObjId(out, gia.Gia_ManCo(out, 1)))
		if left != right {
			t.Fatalf("cones=%v equivalent nodes were not merged", cones)
		}
	}
}

func TestCec2SweepMiterWitness(t *testing.T) {
	for _, cones := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendAnd(g, x, y)
		gia.Gia_ManAppendCo(g, root)
		_, _, cex := cec2SweepEquivalences(g, cec2SweepSettings{finishGraph: true, words: 2, rounds: 1, iterations: 4, conflictLimit: 1000, useCones: cones, isMiter: true})
		if cex == nil || g.SequentialCounterexample() == nil {
			t.Fatal("missing miter witness")
		}
		in := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
		out, err := g.Evaluate(in)
		if err != nil || !out[cex.IPo] {
			t.Fatalf("invalid witness %v: %v", in, err)
		}
	}
}

func TestCec2SweepGraphFinishingIsOptional(t *testing.T) {
	for _, finish := range []bool{false, true} {
		g := gia.Gia_ManStart(8)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendAnd(g, x, y)
		co := gia.Gia_ManAppendCo(g, root)
		out, _, _ := cec2SweepEquivalences(g, cec2SweepSettings{words: 1, iterations: 1, finishGraph: finish})
		if (out != nil) != finish {
			t.Fatalf("finish=%v output=%v", finish, out)
		}
		got := gia.Gia_ObjValue(gia.Gia_ManObj(g, co>>1))
		if (got != -1) != finish {
			t.Fatalf("finish=%v output copy=%d", finish, got)
		}
	}
}
