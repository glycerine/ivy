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
	pars := &Fra_Par_t{NFramesK: 1}
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
	p := Fra_ManStart(original, &Fra_Par_t{NFramesK: 1})
	p.PCla.Reprs[latch.NodeID()] = 0
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

func TestInductionFramesClearDanglingMappings(t *testing.T) {
	original := aig.NewGraph()
	a, b, latch := original.NewInput(), original.NewInput(), original.NewInput()
	dead := original.AND(a, b)
	original.AddOutput(latch)
	aig.Aig_ManSetRegNum(original, 1)
	p := Fra_ManStart(original, &Fra_Par_t{NFramesK: 1})
	Fra_FramesWithClasses(p)
	if got := Fra_ObjFraig(p, dead, 0); got != aig.InvalidEdge {
		t.Fatalf("dangling mapping = %v", got)
	}
}

// fraInd.c unrolls every Aig_ManForEachNode via Aig_And, including EXOR.
func TestInductionFramesTreatNativeExorAsAnd(t *testing.T) {
	source := aig.NewGraph()
	x, y := source.NewInput(), source.NewInput()
	source.NewInput() // latch output
	node := source.NewXor(x, y)
	source.AddOutput(node)
	aig.Aig_ManSetRegNum(source, 1)
	p := Fra_ManStart(source, &Fra_Par_t{NFramesK: 1})
	graph := Fra_FramesWithClasses(p)
	mapped := Fra_ObjFraig(p, node, 0)
	want := graph.AND(Fra_ObjFraig(p, x, 0), Fra_ObjFraig(p, y, 0))
	if mapped != want {
		t.Fatalf("native EXOR unrolled as %v, want AND %v", mapped, want)
	}
}
