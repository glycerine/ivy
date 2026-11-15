package gia

import (
	"fmt"
	"testing"
)

func falseReviewTypedGraph(seed uint32) (*Gia_Man_t, int32) {
	g := Gia_ManStart(128)
	g.EnableMuxes()
	rng := seed
	random := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
	var pool []int32
	for i := 0; i < 4; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	for i := 0; i < 12; i++ {
		a, b, c := int(random())%len(pool), int(random())%len(pool), int(random())%len(pool)
		if a == b {
			b = (b + 1) % len(pool)
		}
		for c == a || c == b {
			c = (c + 1) % len(pool)
		}
		x, y, z := pool[a]^int32(random()&1), pool[b]^int32(random()&1), pool[c]^int32(random()&1)
		var lit int32
		switch i % 4 {
		case 0:
			lit = Gia_ManAppendAnd(g, x, y)
		case 1:
			lit = Gia_ManAppendXorReal(g, x, y)
		case 2:
			lit = Gia_ManAppendMuxReal(g, z, x, y)
		case 3:
			lit = Gia_ManAppendBuf(g, x)
		}
		pool = append(pool, lit)
	}
	root := pool[len(pool)-1] >> 1
	for i := 0; i < 3; i++ {
		Gia_ManAppendCo(g, pool[len(pool)-1-i]^int32(seed&1))
	}
	Gia_ManSetRegNum(g, 1)
	return g, root
}
func TestFalsePathTypedDriverTruth(t *testing.T) {
	for seed := uint32(0); seed < 32; seed++ {
		for _, driver := range []string{"one", "batch", "outputs", "all"} {
			t.Run(fmt.Sprintf("seed%d/%s", seed, driver), func(t *testing.T) {
				g, root := falseReviewTypedGraph(seed)
				defer Gia_ManStop(g)
				Gia_ManLevelNum(g)
				var out *Gia_Man_t
				hashReviewStdout(t, func() {
					switch driver {
					case "one":
						out = Gia_ManCheckOne(g, 0, root, 0, 0, 0)
					case "batch":
						out = Gia_ManCheckFalse2(g, 32, 0, 0, 0)
					case "outputs":
						out = Gia_ManCheckFalse(g, 32, 0, 0, 0)
					case "all":
						out = Gia_ManCheckFalseAll(g, 32, 0, 0, 0)
					}
				})
				if out == nil {
					if driver != "one" {
						t.Fatal("missing result")
					}
					return
				}
				defer Gia_ManStop(out)
				cofReviewCheckRebuild(t, g, out)
			})
		}
	}
}

func falseReviewReconvergentInput() *Gia_Man_t {
	g := Gia_ManStart(32)
	a, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ad := Gia_ManAppendAnd(g, a, d)
	left, right := Gia_ManAppendAnd(g, c^1, ad), Gia_ManAppendAnd(g, c, ad^1)
	xnor := Gia_ManAppendAnd(g, left^1, right^1)
	low, high := Gia_ManAppendAnd(g, c, d^1), Gia_ManAppendAnd(g, d, xnor)
	root := Gia_ManAppendAnd(g, low^1, high^1)
	Gia_ManAppendCo(g, root)
	return g
}
func TestFalsePathBatchPreservesReconvergentInput(t *testing.T) {
	g := falseReviewReconvergentInput()
	defer Gia_ManStop(g)
	out := Gia_ManCheckFalse2(g, 32, 0, 0, 0)
	defer Gia_ManStop(out)
	// The original graph is d ? (a XOR c) : !c. Its timing-path test may
	// reject a transition, but that does not justify replacing the path's d
	// input with zero while keeping reconvergent d side inputs unchanged.
	for bits := 0; bits < 8; bits++ {
		a, c, d := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := !c
		if d {
			want = a != c
		}
		before, after := cofReviewValues(g, bits), cofReviewValues(out, bits)
		if before[Gia_ManCo(g, 0).id] != want {
			t.Fatal("invalid reduced source graph")
		}
		if got := after[Gia_ManCo(out, 0).id]; got != want {
			t.Fatalf("input%d got%t want%t", bits, got, want)
		}
	}
}
