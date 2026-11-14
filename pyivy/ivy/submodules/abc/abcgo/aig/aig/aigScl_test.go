package aig

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

func sclTrace(t *testing.T, p *Graph, stream [][]bool) [][]bool {
	t.Helper()
	regs, pis, pos := sclLayout(p)
	state := make([]bool, regs)
	var trace [][]bool
	for _, input := range stream {
		if len(input) != pis {
			t.Fatal("wrong trace input width")
		}
		all := append(append([]bool(nil), input...), state...)
		out, err := p.Evaluate(all)
		if err != nil {
			t.Fatal(err)
		}
		trace = append(trace, append([]bool(nil), out[:pos]...))
		state = append([]bool(nil), out[pos:]...)
	}
	return trace
}
func sclEqualTrace(t *testing.T, a, b *Graph) {
	t.Helper()
	_, pis, _ := sclLayout(a)
	if _, other, _ := sclLayout(b); other != pis {
		t.Fatal("PI count changed")
	}
	stream := make([][]bool, 24)
	for frame := range stream {
		stream[frame] = make([]bool, pis)
		for i := range stream[frame] {
			stream[frame][i] = (frame>>(i%4))&1 != 0
		}
	}
	left, right := sclTrace(t, a, stream), sclTrace(t, b, stream)
	if len(left) != len(right) {
		t.Fatal("trace length")
	}
	for i := range left {
		if !slices.Equal(left[i], right[i]) {
			t.Fatal("sequential function changed", i, left[i], right[i])
		}
	}
}
func sclGraphFixture() *Graph {
	p := NewGraph()
	pi, live, dependency, dead := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(pi, live))
	p.AddOutput(dependency)
	p.AddOutput(dependency.Not())
	p.AddOutput(p.AND(pi, dead))
	p.nRegs = 3
	return p
}
func TestSequentialRemapSimultaneousAndMetadata(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b.Not()))
	p.AddOutput(a)
	p.nRegs = 1
	q := Aig_ManRemap(p, []Edge{b, a})
	for mask := 0; mask < 4; mask++ {
		x, y := mask&1 != 0, mask&2 != 0
		out, err := q.Evaluate([]bool{x, y})
		if err != nil || out[0] != (y && !x) || out[1] != y {
			t.Fatal(mask, out, err)
		}
	}
	if q.nRegs != 1 || len(q.inputs) != 2 || len(q.outputs) != 2 {
		t.Fatal("remap changed interface")
	}
	tracked := NewGraph()
	pi, r0, r1 := tracked.NewInput(), tracked.NewInput(), tracked.NewInput()
	tracked.AddOutput(tracked.OR(r0, r1))
	tracked.AddOutput(pi)
	tracked.AddOutput(pi)
	tracked.nRegs = 2
	tracked.FlopNumbers = []int32{40, 41}
	tracked.FlopReprs = make([]int32, 0)
	tracked.ClockDomains = [][]int32{{0, 1}}
	remapped := Aig_ManRemap(tracked, []Edge{pi, r0, r0})
	if !slices.Equal(remapped.FlopReprs, []int32{41, 40}) || !slices.Equal(remapped.FlopNumbers, []int32{40, 41}) {
		t.Fatal(remapped.FlopReprs, remapped.FlopNumbers)
	}
	remapped.ClockDomains[0][0] = 1
	if tracked.ClockDomains[0][0] != 0 {
		t.Fatal("metadata alias")
	}
}
func TestSequentialCleanupReachabilityAndBasic(t *testing.T) {
	p := sclGraphFixture()
	p.FlopNumbers = []int32{4, 5, 6}
	p.ClockDomains = [][]int32{{0, 2}, {1}}
	original := Aig_ManDupSimple(p)
	removed := Aig_ManSeqCleanup(p)
	if removed != 1 || p.nRegs != 2 || len(p.inputs) != 3 || len(p.outputs) != 3 || !slices.Equal(p.FlopNumbers, []int32{4, 5}) {
		t.Fatal(removed, p.nRegs, p.FlopNumbers)
	}
	if !slices.Equal(p.ClockDomains[0], []int32{0}) || !slices.Equal(p.ClockDomains[1], []int32{1}) {
		t.Fatal(p.ClockDomains)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	sclEqualTrace(t, original, p)
	basic := sclGraphFixture()
	before := Aig_ManDupSimple(basic)
	Aig_ManSeqCleanupBasic(basic)
	if basic.nRegs != 3 || basic.outputDriver(basic.outputs[len(basic.outputs)-1]) != basic.Const0() {
		t.Fatal("basic cleanup removed terminal or kept dead driver")
	}
	if err := basic.Check(); err != nil {
		t.Fatal(err)
	}
	sclEqualTrace(t, before, basic)
}
func TestSequentialCleanupRecursiveWorklist(t *testing.T) {
	p := sclGraphFixture()
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	p.node(p.inputs[0]).TravID = p.travID
	pending := []Edge{p.outputs[0]}
	for i := 0; i < len(pending); i++ {
		Aig_ManSeqCleanup_rec(p, pending[i], &pending)
	}
	if len(pending) != 3 {
		t.Fatal("next-state worklist did not close", pending)
	}
	if p.node(p.inputs[3]).TravID == p.travID {
		t.Fatal("dead register marked reachable")
	}
}
func TestEqualNextStateRegisterReduction(t *testing.T) {
	p := NewGraph()
	pi, a, b, c := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(a, p.OR(b, c)))
	p.AddOutput(pi)
	p.AddOutput(pi)
	p.AddOutput(pi.Not())
	p.nRegs = 3
	if got := Aig_ManReduceLachesCount(p); got != 1 {
		t.Fatal(got)
	}
	mapping := Aig_ManReduceLachesOnce(p)
	if !slices.Equal(mapping, []Edge{pi, a, a, c}) {
		t.Fatal(mapping)
	}
	original := Aig_ManDupSimple(p)
	result := Aig_ManReduceLaches(p, 0)
	if result.nRegs != 2 {
		t.Fatal(result.nRegs)
	}
	sclEqualTrace(t, original, result)
}
func TestRegisterDependencyComponents(t *testing.T) {
	p := NewGraph()
	p.NewInput()
	b, c, d := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(b)
	p.AddOutput(b)
	p.AddOutput(c)
	p.AddOutput(d)
	p.nRegs = 4
	// 0 -> 1 is one-way, but the historical routine reports weak connectivity.
	components := RegisterComponents(p)
	if len(components) != 3 || !slices.Equal(components[0], []int32{0, 1}) || !slices.Equal(components[1], []int32{2}) || !slices.Equal(components[2], []int32{3}) {
		t.Fatal(components)
	}
}
func TestSequentialSweepDrivers(t *testing.T) {
	for _, mode := range []string{"ternary", "multivalue", "partition", "partition-default"} {
		t.Run(mode, func(t *testing.T) {
			p := NewGraph()
			pi, a, b, zero := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
			p.AddOutput(p.OR(p.AND(pi, zero), p.OR(a, b)))
			p.AddOutput(pi)
			p.AddOutput(pi)
			p.AddOutput(zero)
			p.nRegs = 3
			original := Aig_ManDupSimple(p)
			var result *Graph
			switch mode {
			case "partition":
				p.ClockDomains = [][]int32{{0, 1, 2}}
				result = Aig_ManScl(p, 1, 1, 0, 0, 0, 0, 0)
			case "partition-default":
				result = Aig_ManSclPart(p, 1, 1, 0)
			case "multivalue":
				result = Aig_ManScl(p, 1, 1, 1, 2, 3, 0, 0)
			default:
				result = Aig_ManScl(p, 1, 1, 0, 0, 0, 0, 0)
			}
			if result.nRegs != 1 {
				t.Fatal("failed to reduce registers", result.nRegs)
			}
			if err := result.Check(); err != nil {
				t.Fatal(err)
			}
			sclEqualTrace(t, original, result)
			if len(p.reprs) == 0 {
				t.Fatal("original equivalence classes were not recorded")
			}
		})
	}
}
func TestSequentialDiagnosticAPIs(t *testing.T) {
	p := NewGraph()
	pi, a, b := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(a, b))
	p.AddOutput(pi)
	p.AddOutput(pi)
	p.nRegs = 2

	if Aig_ManCountMergeRegs(p) != 0 {
		t.Fatal("diagnostic return value")
	}
	Aig_ManComputeSccs(p)
	var outputBuffer bytes.Buffer
	writeMergeRegisterStats(&outputBuffer, p)
	writeRegisterComponents(&outputBuffer, p)
	output := outputBuffer.String()
	if !strings.Contains(output, "Regs = 2. Fanins = 1.") || !strings.Contains(output, "SCC #1 contains") {
		t.Fatal(output)
	}
}

func TestSequentialRemapConstantsAndNoRegisters(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	mapped := Aig_ManRemap(p, []Edge{p.Const1(), b.Not()})
	for _, value := range []bool{false, true} {
		out, err := mapped.Evaluate([]bool{false, value})
		if err != nil || out[0] != !value {
			t.Fatal(out, err)
		}
	}
	dangling := p.AND(a, b.Not())
	before := len(p.nodes)
	if removed := Aig_ManSeqCleanup(p); removed != 1 || !p.deleted[dangling.NodeID()] || len(p.inputs) != 2 || len(p.nodes) != before {
		t.Fatal("combinational cleanup changed terminal IDs", removed)
	}
}
func TestSequentialCleanupRetainsMutuallyDependentLiveRegisters(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(a)
	p.AddOutput(b)
	p.AddOutput(a)
	p.nRegs = 2
	original := Aig_ManDupSimple(p)
	if removed := Aig_ManSeqCleanup(p); removed != 0 || p.nRegs != 2 {
		t.Fatal(removed, p.nRegs)
	}
	sclEqualTrace(t, original, p)
}
func TestSequentialInvalidInputPreconditions(t *testing.T) {
	for _, run := range []func(){
		func() { p := NewGraph(); a := p.NewInput(); p.AddOutput(a); Aig_ManRemap(p, nil) },
		func() {
			p := NewGraph()
			a, b := p.NewInput(), p.NewInput()
			p.AddOutput(a)
			Aig_ManRemap(p, []Edge{a, p.AND(a, b)})
		},
		func() { p := NewGraph(); p.nRegs = 1; Aig_ManSeqCleanup(p) },
		func() {
			p := NewGraph()
			a := p.NewInput()
			p.AddOutput(a)
			p.nRegs = 1
			p.FlopNumbers = []int32{}
			Aig_ManSeqCleanup(p)
		},
		func() {
			p := NewGraph()
			a := p.NewInput()
			p.AddOutput(a)
			p.nRegs = 1
			p.ClockDomains = [][]int32{{2}}
			Aig_ManSeqCleanup(p)
		},
		func() { Aig_ManReduceLachesCount(NewGraph()) },
		func() {
			p := NewGraph()
			a := p.NewInput()
			p.AddOutput(a)
			p.nRegs = 1
			p.FlopNumbers = []int32{0}
			Aig_ManScl(p, 0, 0, 0, 0, 0, 0, 0)
		},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid sequential input accepted")
				}
			}()
			run()
		}()
	}
}

func TestRegisterReductionFixedPointAndOriginalFlopNumbers(t *testing.T) {
	build := func() *Graph {
		p := NewGraph()
		pi, a, b, c, d := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
		p.AddOutput(p.OR(a, p.OR(c, d)))
		p.AddOutput(pi)
		p.AddOutput(pi)
		p.AddOutput(a)
		p.AddOutput(b)
		p.nRegs = 4
		return p
	}
	original := build()
	p := build()
	p.FlopNumbers = []int32{10, 11, 12, 13}
	p.FlopReprs = make([]int32, 0)
	result := Aig_ManReduceLaches(p, 0)
	if result.nRegs != 2 || !slices.Equal(result.FlopNumbers, []int32{10, 12}) || !slices.Equal(result.FlopReprs, []int32{11, 10, 13, 12}) {
		t.Fatal(result.nRegs, result.FlopNumbers, result.FlopReprs)
	}
	sclEqualTrace(t, original, result)
	swept := Aig_ManScl(build(), 0, 1, 0, 0, 0, 0, 0)
	if swept.nRegs != 2 {
		t.Fatal("full sweep did not reach equality fixed point", swept.nRegs)
	}
	sclEqualTrace(t, original, swept)
}
func TestSequentialCleanupRejectsBadClockMetadataWithoutMutation(t *testing.T) {
	p := sclGraphFixture()
	p.ClockDomains = [][]int32{{3}}
	beforeInputs, beforeOutputs := p.Inputs(), p.Outputs()
	beforeRegs := p.nRegs
	defer func() {
		if recover() == nil {
			t.Error("bad clock domain accepted")
		}
		if !slices.Equal(p.inputs, beforeInputs) || !slices.Equal(p.outputs, beforeOutputs) || p.nRegs != beforeRegs {
			t.Error("graph mutated before rejecting metadata")
		}
	}()
	Aig_ManSeqCleanup(p)
}

func TestSequentialCleanupRetainsStandaloneBufferStorage(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	gate := p.AND(a, b)
	buffer := Aig_ObjCreate(p, &Node{Kind: Buffer, Fanins: [2]Edge{gate, InvalidEdge}, Level: p.node(gate).Level})
	if count := Aig_ManSeqCleanup(p); count != 0 || p.deleted[gate.NodeID()] || p.deleted[buffer.NodeID()] {
		t.Fatal(count, p.deleted)
	}
}

func TestSequentialSweepIdempotenceAndEmptyDiagnostics(t *testing.T) {
	p := NewGraph()
	pi, a, b := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(a, b))
	p.AddOutput(pi)
	p.AddOutput(pi)
	p.nRegs = 2
	once := Aig_ManScl(p, 1, 1, 0, 0, 0, 0, 0)
	original := Aig_ManDupSimple(once)
	twice := Aig_ManScl(once, 1, 1, 0, 0, 0, 0, 0)
	if twice.nRegs != 1 {
		t.Fatal(twice.nRegs)
	}
	sclEqualTrace(t, original, twice)
	var text bytes.Buffer
	writeRegisterComponents(&text, NewGraph())
	if text.String() != "The network is combinational.\n" {
		t.Fatal(text.String())
	}
}

func TestTrackedRemapRejectsUnrepresentableZeroStateClasses(t *testing.T) {
	for _, targetKind := range []string{"one", "primary-input", "complemented-register"} {
		t.Run(targetKind, func(t *testing.T) {
			p := NewGraph()
			pi, lo := p.NewInput(), p.NewInput()
			p.AddOutput(lo)
			p.AddOutput(lo)
			p.nRegs = 1
			p.FlopNumbers = []int32{0}
			p.FlopReprs = make([]int32, 0)
			target := p.Const1()
			switch targetKind {
			case "primary-input":
				target = pi
			case "complemented-register":
				target = lo.Not()
			}
			defer func() {
				if recover() == nil {
					t.Error("unrepresentable register class accepted")
				}
				if p.copies != nil {
					t.Error("invalid mapping modified source scratch before rejection")
				}
			}()
			Aig_ManRemap(p, []Edge{pi, target})
		})
	}
}

func TestBasicCleanupHonorsComplementedOutputEdge(t *testing.T) {
	p := NewGraph()
	pi, lo := p.NewInput(), p.NewInput()
	p.AddOutput(pi)
	p.AddOutput(lo)
	p.nRegs = 1
	p.outputs[1] = p.outputs[1].Not()
	Aig_ManSeqCleanupBasic(p)
	if driver := p.outputDriver(p.outputs[1]); driver != p.Const0() {
		t.Fatal("inactive register was not tied to logical zero", driver)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
