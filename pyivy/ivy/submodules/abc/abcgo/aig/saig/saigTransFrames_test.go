package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFramesNonInitial(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.XOR(pi, lo)
	g.AddOutput(lo)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	q := Saig_ManFramesNonInitial(g, 2)
	defer Saig_ManStopMap1(g)
	if q == nil || q.InputCount() != 3 || q.OutputCount() != 3 {
		t.Fatalf("frame graph %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	if Saig_ManGetMap1(g, pi, 0) != int32(q.Inputs()[1].NodeID()) || Saig_ManGetMap1(g, pi, 1) != int32(q.Inputs()[2].NodeID()) || Saig_ManGetMap1(g, lo, 0) != int32(q.Inputs()[0].NodeID()) {
		t.Fatal("incorrect input maps")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != in[0] || out[1] != (in[0] != in[1]) || out[2] != (in[0] != in[1] != in[2]) {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}

func TestFramesInitialMapped(t *testing.T) {
	for _, init := range []int32{0, 1} {
		for _, mapped := range []bool{false, true} {
			g := aig.NewGraph()
			pi, lo := g.NewInput(), g.NewInput()
			g.AddOutput(lo)
			g.AddOutput(g.XOR(pi, lo))
			aig.Aig_ManSetRegNum(g, 1)
			if mapped {
				Saig_ManStartMap2(g, 2)
				for f := int32(0); f < 2; f++ {
					for _, e := range g.LiveEdges() {
						Saig_ManSetMap2(g, e, f, e, f)
					}
				}
			}
			q := Saig_ManFramesInitialMapped(g, 2, 3, init)
			if mapped {
				Saig_ManStopMap2(g)
			}
			if q == nil {
				t.Fatal("missing graph")
			}
			if err := q.Check(); err != nil {
				t.Fatal(err)
			}
			wantInputs, wantOutputs, wantRegs := 3, 3, int32(0)
			if init == 0 {
				wantInputs, wantOutputs, wantRegs = 4, 4, 1
			}
			if q.InputCount() != wantInputs || q.OutputCount() != wantOutputs || q.RegisterCount() != wantRegs {
				t.Fatalf("init=%d mapped=%v inputs=%d outputs=%d regs=%d", init, mapped, q.InputCount(), q.OutputCount(), q.RegisterCount())
			}
			for bits := 0; bits < (1 << wantInputs); bits++ {
				in := make([]bool, wantInputs)
				for i := range in {
					in[i] = bits&(1<<i) != 0
				}
				out, err := q.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				state := false
				if init == 0 {
					state = in[3]
				}
				for f := 0; f < 3; f++ {
					if out[f] != state {
						t.Fatalf("init=%d mapped=%v in=%v out=%v", init, mapped, in, out)
					}
					state = state != in[f]
				}
				if init == 0 && out[3] != state {
					t.Fatalf("incorrect next state: %v", out)
				}
			}
		}
	}
}
