package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func phaseFixture() *aig.Graph {
	graph := aig.NewGraph()
	x, clock, data := graph.NewInput(), graph.NewInput(), graph.NewInput()
	graph.AddOutput(data)
	graph.AddOutput(graph.AND(x, clock))
	graph.AddOutput(clock.Not())
	graph.AddOutput(x)
	aig.Aig_ManSetRegNum(graph, 2)
	return graph
}
func TestPhaseTernaryCycleAndCandidates(t *testing.T) {
	graph := phaseFixture()
	p := Saig_ManReachableTernary(graph, nil, 0)
	if p == nil || p.nPrefix != 1 || p.nCycle != 2 {
		t.Fatalf("ternary recurrence %+v", p)
	}
	if count := Saig_TsiCountNonXValuedRegisters(p, 0); count != 1 || vec.Vec_IntEntry(p.vNonXRegs, 0) != 0 {
		t.Fatalf("candidate count %d", count)
	}
	if count := Saig_ManFindRegisters(p, 2, 0, 0); count != 1 {
		t.Fatalf("clock period not found %d", count)
	}
	if got := Saig_ManPhaseFrameNum(graph, nil); got != 2 {
		t.Fatalf("frame count %d", got)
	}
}
func TestPhaseAbstractionPreservesGroupedTraces(t *testing.T) {
	source := phaseFixture()
	abstracted := Saig_ManPhaseAbstract(source, nil, 2, 0, 0, 0, 0)
	if abstracted == nil {
		t.Fatal("clock phase abstraction refused")
	}
	if aig.Aig_ManRegNum(abstracted) >= aig.Aig_ManRegNum(source) {
		t.Fatal("clock register not eliminated")
	}
	for pattern := 0; pattern < 256; pattern++ {
		originalState := []bool{false, false}
		abstractState := make([]bool, aig.Aig_ManRegNum(abstracted))
		for step := 0; step < 4; step++ {
			x0, x1 := pattern&(1<<uint(2*step)) != 0, pattern&(1<<uint(2*step+1)) != 0
			expected := make([]bool, 0, 4)
			for _, x := range []bool{x0, x1} {
				outputs, err := source.Evaluate(append([]bool{x}, originalState...))
				if err != nil {
					t.Fatal(err)
				}
				expected = append(expected, outputs[:2]...)
				copy(originalState, outputs[2:])
			}
			outputs, err := abstracted.Evaluate(append([]bool{x0, x1}, abstractState...))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expected, outputs[:4]) {
				t.Fatalf("grouped trace input=%d step=%d got=%v expected=%v", pattern, step, outputs, expected)
			}
			copy(abstractState, outputs[4:])
		}
	}
}
func TestPhaseTransientResetsPerRegister(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	graph.AddOutput(a)
	graph.AddOutput(b)
	aig.Aig_ManSetRegNum(graph, 2)
	p := Saig_TsiStart(graph)
	for _, values := range [][2]int32{{SAIG_XVS0, SAIG_XVS1}, {SAIG_XVS1, SAIG_XVS1}, {SAIG_XVS1, SAIG_XVS1}} {
		state := Saig_TsiStateNew(p)
		for reg, value := range values {
			phaseSet(state, reg, value)
		}
	}
	if got := vec.Vec_IntArray(Saig_TsiComputeTransient(p, 2)); !reflect.DeepEqual(got, []int32{1, 1}) {
		t.Fatalf("transient histogram %v", got)
	}
}
func TestPhaseCounterexampleTranslation(t *testing.T) {
	source := phaseFixture()
	cex := util.Abc_CexAlloc(1, 2, 3)
	cex.IFrame = 2
	cex.IPo = 3
	for i := int32(0); i < 6; i++ {
		if i%2 == 1 {
			bit := cex.NRegs + i
			cex.PData[bit>>5] |= 1 << uint(bit&31)
		}
	}
	translated := Saig_PhaseTranslateCex(source, cex)
	if translated == nil || translated.IFrame != 5 || translated.IPo != 1 || translated.NRegs != 2 {
		t.Fatalf("translated witness %+v", translated)
	}
	for i := int32(0); i < 6; i++ {
		if got := translated.PData[(translated.NRegs+i)>>5]&(1<<uint((translated.NRegs+i)&31)) != 0; got != (i%2 == 1) {
			t.Fatal("translated inputs changed")
		}
	}
}
