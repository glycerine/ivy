package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestInductionPreservesReachableOutput(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(input)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 2)
	pars := &Fra_Ssw_t{nFramesK: 1}
	got := Fra_FraigInduction(g, pars)
	if got == nil {
		t.Fatal("induction returned no graph")
	}
	for mask := 0; mask < 8; mask++ {
		srcRegs := make([]bool, g.RegisterCount())
		dstRegs := make([]bool, got.RegisterCount())
		for frame := 0; frame < 3; frame++ {
			in := mask&(1<<frame) != 0
			old, err := g.Evaluate(append([]bool{in}, srcRegs...))
			if err != nil {
				t.Fatal(err)
			}
			out, err := got.Evaluate(append([]bool{in}, dstRegs...))
			if err != nil {
				t.Fatal(err)
			}
			if old[0] != out[0] {
				t.Fatalf("mask=%d frame=%d original=%v reduced=%v", mask, frame, old, out)
			}
			copy(srcRegs, old[len(old)-len(srcRegs):])
			copy(dstRegs, out[len(out)-len(dstRegs):])
		}
	}
}

func TestPartitionProjectsKnownOneHotGroups(t *testing.T) {
	g := aig.NewGraph()
	pi, a, b, c := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(pi)
	g.AddOutput(pi)
	g.AddOutput(c)
	aig.Aig_ManSetRegNum(g, 3)
	g.OneHotGroups = [][]int32{{0, 1, 2}}
	part := aig.Aig_ManRegCreatePart(g, []int32{0, 1}, nil, nil, nil)
	fraProjectPartitionOneHots(g, part, 0)
	if len(part.OneHotInputGroups) != 1 || len(part.OneHotInputGroups[0]) < 2 {
		t.Fatalf("projected one-hot groups = %v", part.OneHotInputGroups)
	}
	group := part.OneHotInputGroups[0]
	if group[0] < 0 || int(group[0]) >= part.InputCount() || group[1] < 0 || int(group[1]) >= part.InputCount() || group[0] == group[1] {
		t.Fatalf("invalid projected inputs %v", group)
	}
}

func TestInductionOptionsPreserveReachableOutput(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		rewrite, oneHot, imps int32
	}{
		{"rewrite", 1, 0, 0}, {"one-hot", 0, 1, 0}, {"implications", 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.XOR(a, b))
			g.AddOutput(input)
			g.AddOutput(input)
			aig.Aig_ManSetRegNum(g, 2)
			result := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: 1, fRewrite: tc.rewrite, fUse1Hot: tc.oneHot, fUseImps: tc.imps, nMaxImps: 16})
			if result == nil {
				t.Fatal("no induction result")
			}
			for mask := 0; mask < 8; mask++ {
				src, dst := make([]bool, g.RegisterCount()), make([]bool, result.RegisterCount())
				for frame := 0; frame < 3; frame++ {
					input := mask&(1<<frame) != 0
					a, err := g.Evaluate(append([]bool{input}, src...))
					if err != nil {
						t.Fatal(err)
					}
					b, err := result.Evaluate(append([]bool{input}, dst...))
					if err != nil {
						t.Fatal(err)
					}
					if a[0] != b[0] {
						t.Fatalf("mask %d frame %d outputs %v/%v", mask, frame, a, b)
					}
					copy(src, a[len(a)-len(src):])
					copy(dst, b[len(b)-len(dst):])
				}
			}
		})
	}
}

func TestInductionExhaustiveOneRegisterTransitions(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		g := aig.NewGraph()
		input, latch := g.NewInput(), g.NewInput()
		next := g.Const0()
		for row := 0; row < 4; row++ {
			if truth&(1<<row) == 0 {
				continue
			}
			a, b := input, latch
			if row&1 == 0 {
				a = a.Not()
			}
			if row&2 == 0 {
				b = b.Not()
			}
			next = g.OR(next, g.AND(a, b))
		}
		g.AddOutput(g.XOR(input, latch))
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		result := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: 1})
		if result == nil {
			t.Fatalf("transition %04b returned no graph", truth)
		}
		for sequence := 0; sequence < 16; sequence++ {
			src, dst := make([]bool, g.RegisterCount()), make([]bool, result.RegisterCount())
			for frame := 0; frame < 4; frame++ {
				value := sequence&(1<<frame) != 0
				a, err := g.Evaluate(append([]bool{value}, src...))
				if err != nil {
					t.Fatal(err)
				}
				b, err := result.Evaluate(append([]bool{value}, dst...))
				if err != nil {
					t.Fatal(err)
				}
				if a[0] != b[0] {
					t.Fatalf("transition %04b sequence %04b frame %d outputs %v/%v", truth, sequence, frame, a, b)
				}
				copy(src, a[len(a)-len(src):])
				copy(dst, b[len(b)-len(dst):])
			}
		}
	}
}

func TestInductionPrefixAndPartitionPreserveOutputs(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		frames, prefix, partition int32
	}{
		{"two-frames", 2, 0, 0}, {"prefix", 1, 1, 0}, {"partitions", 1, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.XOR(a, b))
			g.AddOutput(g.AND(a, input))
			g.AddOutput(input)
			g.AddOutput(input)
			aig.Aig_ManSetRegNum(g, 2)
			if tc.partition != 0 {
				g.OneHotGroups = [][]int32{{0, 1}}
			}
			out := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: tc.frames, nFramesP: tc.prefix, nPartSize: tc.partition})
			if out == nil {
				t.Fatal("induction returned nil")
			}
			for sequence := 0; sequence < 16; sequence++ {
				src, dst := make([]bool, g.RegisterCount()), make([]bool, out.RegisterCount())
				for frame := 0; frame < 4; frame++ {
					input := sequence&(1<<frame) != 0
					left, err := g.Evaluate(append([]bool{input}, src...))
					if err != nil {
						t.Fatal(err)
					}
					right, err := out.Evaluate(append([]bool{input}, dst...))
					if err != nil {
						t.Fatal(err)
					}
					for i := 0; i < 2; i++ {
						if left[i] != right[i] {
							t.Fatalf("sequence=%04b frame=%d output=%d original=%v reduced=%v", sequence, frame, i, left, right)
						}
					}
					copy(src, left[len(left)-len(src):])
					copy(dst, right[len(right)-len(dst):])
				}
			}
		})
	}
}

func TestInductionTimeoutReturnsNoResult(t *testing.T) {
	g := aig.NewGraph()
	input, latch := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(input, latch))
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Fra_Ssw_t{nFramesK: 1, TimeLimit: 1e-12, fSilent: 1}
	if result := Fra_FraigInduction(g, pars); result != nil {
		t.Fatal("expected timed-out induction to return no graph")
	}
	if pars.nIters != -1 {
		t.Fatalf("iteration count=%d", pars.nIters)
	}
}

func TestInductionWithNontrivialOneHotCandidate(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(input)
	g.AddOutput(input.Not())
	aig.Aig_ManSetRegNum(g, 2)
	sim := Fra_SmlSimulateSeq(g, 0, 32, 1, 1)
	probe := &Fra_Man_t{PManAig: g}
	if Fra_OneHotCount(probe, Fra_OneHotCompute(probe, sim)) == 0 {
		t.Fatal("fixture produced no one-hot candidate")
	}
	result := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: 1, fUse1Hot: 1})
	if result == nil {
		t.Fatal("one-hot induction returned no result")
	}
	for sequence := 0; sequence < 16; sequence++ {
		src, dst := make([]bool, 2), make([]bool, result.RegisterCount())
		for frame := 0; frame < 4; frame++ {
			value := sequence&(1<<frame) != 0
			a, err := g.Evaluate(append([]bool{value}, src...))
			if err != nil {
				t.Fatal(err)
			}
			b, err := result.Evaluate(append([]bool{value}, dst...))
			if err != nil {
				t.Fatal(err)
			}
			if a[0] != b[0] {
				t.Fatalf("sequence %04b frame %d outputs %v/%v", sequence, frame, a, b)
			}
			copy(src, a[len(a)-len(src):])
			copy(dst, b[len(b)-len(dst):])
		}
	}
}

func TestInductionWithNontrivialImplicationCandidate(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b.Not()))
	g.AddOutput(input)
	g.AddOutput(g.OR(input, a))
	aig.Aig_ManSetRegNum(g, 2)
	probe := Fra_ManStart(g, &Fra_Par_t{NFramesK: 1})
	imps := Fra_ImpDerive(probe, 100, 20, 0)
	found := false
	for _, imp := range vec.Vec_IntArray(imps) {
		if Fra_ImpLeft(imp) == int32(a.NodeID()) && Fra_ImpRight(imp) == int32(b.NodeID()) {
			found = true
		}
	}
	if !found {
		t.Fatalf("fixture did not produce a => b: %v", vec.Vec_IntArray(imps))
	}
	result := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: 1, fUseImps: 1, nMaxImps: 20})
	if result == nil {
		t.Fatal("implication induction returned no result")
	}
	for sequence := 0; sequence < 16; sequence++ {
		src, dst := make([]bool, 2), make([]bool, result.RegisterCount())
		for frame := 0; frame < 4; frame++ {
			value := sequence&(1<<frame) != 0
			a, err := g.Evaluate(append([]bool{value}, src...))
			if err != nil {
				t.Fatal(err)
			}
			b, err := result.Evaluate(append([]bool{value}, dst...))
			if err != nil {
				t.Fatal(err)
			}
			if a[0] != b[0] {
				t.Fatalf("sequence %04b frame %d outputs %v/%v", sequence, frame, a, b)
			}
			copy(src, a[len(a)-len(src):])
			copy(dst, b[len(b)-len(dst):])
		}
	}
}
