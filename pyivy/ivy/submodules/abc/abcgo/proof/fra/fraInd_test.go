package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
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
