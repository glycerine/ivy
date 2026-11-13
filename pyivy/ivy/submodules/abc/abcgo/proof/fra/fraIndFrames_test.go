package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInductionFramesTransferLatch(t *testing.T) {
	original := aig.NewGraph()
	input, latch := original.NewInput(), original.NewInput()
	original.AddOutput(input)
	aig.Aig_ManSetRegNum(original, 1)
	pars := &Fra_Par_t{nFramesK: 1}
	p := Fra_ManStart(original, pars)
	graph := Fra_FramesWithClasses(p)
	if graph.InputCount() != 3 || graph.OutputCount() != 1 || graph.RegisterCount() != 1 || graph.Assertions != 0 {
		t.Fatalf("interface %d/%d regs=%d asserts=%d", graph.InputCount(), graph.OutputCount(), graph.RegisterCount(), graph.Assertions)
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != inputs[0] {
			t.Fatalf("inputs %v output %v", inputs, out)
		}
	}
	_ = latch
}

func TestInductionFramesEmitClassConstraint(t *testing.T) {
	original := aig.NewGraph()
	input, latch := original.NewInput(), original.NewInput()
	original.AddOutput(input)
	aig.Aig_ManSetRegNum(original, 1)
	p := Fra_ManStart(original, &Fra_Par_t{nFramesK: 1})
	p.pCla.reprs[latch.NodeID()] = 0
	graph := Fra_FramesWithClasses(p)
	if graph.Assertions != 1 || graph.OutputCount() != 2 {
		t.Fatalf("assertions=%d outputs=%d", graph.Assertions, graph.OutputCount())
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] == inputs[2] {
			t.Fatalf("constraint should require a zero latch: %v", inputs)
		}
	}
}
