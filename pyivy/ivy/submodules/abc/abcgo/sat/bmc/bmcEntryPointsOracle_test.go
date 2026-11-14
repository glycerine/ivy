package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// A latch initialized to zero whose next state is the primary input.
func entryPointGraph() *aig.Graph {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}

func TestBoundedCheckingEntryPoints(t *testing.T) {
	entries := []struct {
		name string
		run  func(*gia.Gia_Man_t, *Bmc_AndPar_t) int32
	}{
		{"bmcg-one", Bmcg_ManPerformOne}, {"bmcg", Bmcg_ManPerform},
		{"bmcs-one", Bmcs_ManPerformOne}, {"bmcs", Bmcs_ManPerform},
	}
	for _, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			for _, bound := range []int32{1, 2} {
				g := gia.Gia_ManFromAig(entryPointGraph())
				p := &Bmc_AndPar_t{NFramesMax: bound}
				status := entry.run(g, p)
				if bound == 1 && status != 1 {
					t.Fatalf("bound 1 status %d", status)
				}
				if bound == 2 {
					cex := g.SequentialCounterexample()
					if status != 0 || cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
						t.Fatalf("status=%d witness=%+v", status, cex)
					}
				}
			}
		})
	}
}

func TestInitialStateAndChainEntryPoints(t *testing.T) {
	identity := stateRegister(true)
	g := gia.Gia_ManFromAig(identity)
	for _, tc := range []struct {
		initial int32
		want    int32
	}{{0, 1}, {1, 0}, {2, 0}} {
		init := vec.Vec_IntAllocArrayCopy([]int32{tc.initial}, 1)
		if got := Gia_ManBmciTest(g, init, 1, 1, 0, 0, 0); got != tc.want {
			t.Errorf("initial=%d got %d want %d", tc.initial, got, tc.want)
		}
	}
	if got := Gia_ManBmciTest(g, vec.Vec_IntAlloc(0), 1, 1, 0, 0, 0); got != -1 {
		t.Fatalf("invalid initial length returned %d", got)
	}
	witness := Bmc_ChainFailOneOutput(gia.Gia_ManFromAig(entryPointGraph()), 2, 0, 0, 0)
	if witness == nil || witness.IFrame != 1 || witness.IPo != 0 {
		t.Fatalf("chain witness %+v", witness)
	}
}

func TestUnrollEntryPointsMatchExplicitFrames(t *testing.T) {
	source := entryPointGraph()
	manager := gia.Gia_ManFromAig(source)
	u := Unr_ManUnrollStart(manager, 0)
	if u == nil {
		t.Fatal("nil unroll manager")
	}
	defer Unr_ManFree(u)
	for frames := int32(1); frames <= 3; frames++ {
		graphs := []*gia.Gia_Man_t{Unr_ManUnroll(manager, frames), Unr_ManUnrollSimple(manager, frames), Unr_ManUnrollFrame(u, frames-1), gia.Gia_ManFromAig(Saig_ManFramesBmc(source, frames))}
		for index, g := range graphs {
			if g == nil {
				t.Fatalf("frames=%d entry=%d nil", frames, index)
			}
			inputs := len(g.InputLiterals())
			if inputs != int(frames) {
				t.Fatalf("frames=%d entry=%d inputs=%d", frames, index, inputs)
			}
			for bits := 0; bits < 1<<inputs; bits++ {
				in := make([]bool, inputs)
				for i := range in {
					in[i] = bits&(1<<i) != 0
				}
				out, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for f := 0; f < int(frames); f++ {
					want := f > 0 && in[f-1]
					if out[f] != want {
						t.Fatalf("frames=%d entry=%d input=%v outputs=%v frame=%d want=%v", frames, index, in, out, f, want)
					}
				}
			}
		}
	}
}
