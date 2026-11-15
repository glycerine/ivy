package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCubeReachabilityTypedInitialViolation(t *testing.T) {
	for _, kind := range []string{"ordinary", "xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			defer Gia_ManStop(g)
			a, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			out := a
			switch kind {
			case "xor":
				out = Gia_ManAppendXorReal(g, a, r)
			case "mux":
				g.EnableMuxes()
				out = Gia_ManAppendMuxReal(g, a, 1, r)
			}
			Gia_ManAppendCo(g, out)
			Gia_ManAppendCo(g, r)
			Gia_ManSetRegNum(g, 1)
			var got int32
			hashReviewStdout(t, func() { got = Gia_ManArePerform(g, 100, 1, 0) })
			if got != 0 {
				t.Fatalf("initial input=true violation returned%d, want0", got)
			}
			if g.pCexSeq == nil || g.pCexSeq.IFrame != 0 || Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
				t.Fatalf("invalid initial counterexample: %+v", g.pCexSeq)
			}
		})
	}
}

func TestCubeReachabilityMuxControlFanout(t *testing.T) {
	g := Gia_ManStart(16)
	defer Gia_ManStop(g)
	c, a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	mux := Gia_ManAppendMuxReal(g, c, a, b)
	co := Gia_ManAppendCo(g, mux)
	cones := Gia_ManDeriveCiTfo(g)
	got := vec.Vec_IntArray(vec.Vec_PtrEntry(cones, 0).(*Vec_Int_t))
	want := []int32{c >> 1, mux >> 1, co >> 1}
	if len(got) != len(want) {
		t.Fatalf("control fanout%v want%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("control fanout%v want%v", got, want)
		}
	}
	for _, obj := range g.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("fanout traversal retained marks")
		}
	}
}

func TestCubeReachabilityTypedSequentialAgreement(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		for _, safe := range []bool{false, true} {
			t.Run(fmt.Sprintf("phase%d/safe%t", phase, safe), func(t *testing.T) {
				g := eraReviewTypedMachine(phase)
				defer Gia_ManStop(g)
				if safe {
					Gia_ManPatchCoDriver(g, 0, 0)
					Gia_ManPatchCoDriver(g, 1, 0)
				}
				expanded := Gia_ManDupNoMuxes(g, 1)
				defer Gia_ManStop(expanded)
				for miter := int32(0); miter < 2; miter++ {
					want := int32(1)
					if miter != 0 && !safe {
						want = 0
					}
					for _, graph := range []*Gia_Man_t{g, expanded} {
						var got int32
						hashReviewStdout(t, func() { got = Gia_ManArePerform(graph, 100, miter, 0) })
						if got != want {
							t.Fatalf("typed%t miter%d got%d want%d", graph == g, miter, got, want)
						}
						if got == 0 && (graph.pCexSeq == nil || Gia_ManVerifyCex(graph, graph.pCexSeq, 0) == 0) {
							t.Fatal("invalid sequential counterexample")
						}
						if got != 0 && graph.pCexSeq != nil {
							t.Fatal("safe run retained previous failure trace")
						}
					}
				}
			})
		}
	}
}
