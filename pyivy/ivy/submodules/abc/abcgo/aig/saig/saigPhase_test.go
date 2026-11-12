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

func TestPhaseRetiresLongBinaryCounter(t *testing.T) {
	graph := aig.NewGraph()
	graph.NewInput()
	state := make([]aig.Edge, 13)
	for i := range state {
		state[i] = graph.NewInput()
	}
	graph.AddOutput(state[12])
	carry := graph.Const1()
	for _, bit := range state {
		graph.AddOutput(graph.NewXor(bit, carry))
		carry = graph.AND(carry, bit)
	}
	aig.Aig_ManSetRegNum(graph, 13)
	p := Saig_ManReachableTernary(graph, nil, 0)
	if p == nil || p.nPrefix < 3000 || p.nCycle != 1 || len(p.states) > 4000 {
		t.Fatalf("long-counter retirement %+v", p)
	}
	if n := Saig_TsiCountNonXValuedRegisters(p, p.nPrefix); n != 0 {
		t.Fatalf("retired counter retained %d binary bits", n)
	}
}
func TestPhaseConstraintOrderingAndTranslation(t *testing.T) {
	graph := phaseFixture()
	aig.Aig_ManSetConstrNum(graph, 1)
	abstracted := Saig_ManPhaseAbstract(graph, nil, 2, 0, 0, 0, 0)
	if abstracted == nil || aig.Aig_ManConstrNum(abstracted) != 2 {
		t.Fatal("constraint count not expanded")
	}
	// The second property comes from microframe 1. Constraints follow all
	// property frames, rather than being interleaved with each property.
	cex := util.Abc_CexAlloc(aig.Aig_ManRegNum(abstracted), 2, 1)
	cex.IPo = 1
	translated := Saig_PhaseTranslateCex(graph, cex)
	if translated == nil || translated.IFrame != 1 || translated.IPo != 0 {
		t.Fatalf("property coordinate translation %+v", translated)
	}
	cex.IPo = 3
	translated = Saig_PhaseTranslateCex(graph, cex)
	if translated == nil || translated.IFrame != 1 || translated.IPo != 1 {
		t.Fatalf("constraint coordinate translation %+v", translated)
	}
}

func TestPhaseIgnoreConstantRegisters(t *testing.T) {
	graph := aig.NewGraph()
	graph.NewInput()
	clock, constant := graph.NewInput(), graph.NewInput()
	graph.AddOutput(clock)
	graph.AddOutput(clock.Not())
	graph.AddOutput(constant)
	aig.Aig_ManSetRegNum(graph, 2)
	p := Saig_ManReachableTernary(graph, nil, 0)
	if Saig_TsiCountNonXValuedRegisters(p, 0) != 2 {
		t.Fatal("binary candidates missing")
	}
	if n := Saig_ManFindRegisters(p, 2, 1, 0); n != 1 || vec.Vec_IntEntry(p.vNonXRegs, 0) != 0 {
		t.Fatalf("ignore-constant selection %v", vec.Vec_IntArray(p.vNonXRegs))
	}
}
func TestPhaseRejectsTruncatedCounterexample(t *testing.T) {
	source := phaseFixture()
	cex := &util.Abc_Cex_t{NRegs: 1, NPis: 2, NBits: 7, IFrame: 2, IPo: 0}
	if got := Saig_PhaseTranslateCex(source, cex); got != nil {
		t.Fatal("accepted missing witness bits")
	}
}
