package aig

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMultivalueFrameAllocatesSymbolsInSourceObjectOrder(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	latePI, lo := g.NewInput(), g.NewInput()
	g.AddOutput(gate)
	g.AddOutput(latePI)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	p.values[lo.NodeID()] = 1
	Aig_MvSimulateFrame(p, 1, 0)
	// Original C visits a,b,gate,latePI in object-ID order. The gate must
	// consume its symbolic literal before the later PI receives one.
	if p.values[gate.NodeID()] != 6 || p.values[latePI.NodeID()] != 8 || p.values[lo.NodeID()] != 8 {
		t.Fatalf("gate/late PI/LO = %d/%d/%d, want 6/8/8", p.values[gate.NodeID()], p.values[latePI.NodeID()], p.values[lo.NodeID()])
	}
}

func TestMultivalueTraceAndMappingAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "multivalue_trace_compare.c", "aigMem.c")
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	late := g.NewInput()
	los := make([]Edge, 5)
	for i := range los {
		los[i] = g.NewInput()
	}
	g.AddOutput(gate)
	g.AddOutput(los[0])
	g.AddOutput(los[1].Not())
	g.AddOutput(late)
	g.AddOutput(late)
	g.AddOutput(gate)
	g.nRegs = 5
	p := Aig_MvManStart(g, 3)
	for _, e := range los {
		p.values[e.NodeID()] = 1
	}
	var got strings.Builder
	fmt.Fprintf(&got, "initial %d\n", Aig_MvSaveState(p))
	for f := 0; f < 8; f++ {
		first := int32(0)
		if f < 2 {
			first = 1
		}
		Aig_MvSimulateFrame(p, first, 0)
		state := Aig_MvSaveState(p)
		fmt.Fprintf(&got, "frame %d state %d objects %d:", f, state, p.nextLiteral/2)
		for _, v := range p.values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
		if state >= 0 {
			break
		}
	}
	printInts := func(name string, values []int32) {
		fmt.Fprintf(&got, "%s:", name)
		for _, v := range values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
	}
	var binary []int32
	constants := Aig_MvManFindConstBinaryFlops(p, &binary)
	printInts("undefined", Aig_MvManFindXFlops(p))
	printInts("constants", constants)
	printInts("binary", binary)
	printInts("oscillators", Aig_MvManFindOscilators(p, nil))
	printInts("skip", Aig_MvManCreateNextSkip(p))
	printMap := func(name string, values []Edge) {
		fmt.Fprintf(&got, "%s:", name)
		for _, e := range values {
			fmt.Fprintf(&got, " %d", e)
		}
		got.WriteByte('\n')
	}
	printMap("mapping", Aig_MvManDeriveMap(p, 0))
	printMap("driver", Aig_MvManSimulate(g, 2, 3, 0, 0))
	if got.String() != out {
		t.Fatalf("Go trace:\n%sC trace:\n%s", got.String(), out)
	}
}

func TestMultivalueFrameSupportsForwardBufferDependencies(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	early := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	later := g.AND(a, b)
	Aig_ObjConnect(g, early, later, InvalidEdge)
	lo := g.NewInput()
	po := g.AddOutput(early.Not())
	g.AddOutput(early)
	g.nRegs = 1
	before := g.Nodes()
	p := Aig_MvManStart(g, 2)
	p.values[lo.NodeID()] = 1
	Aig_MvSimulateFrame(p, 1, 0)
	if p.values[later.NodeID()] != 6 || p.values[early.NodeID()] != 6 || p.values[po.NodeID()] != 7 || p.values[lo.NodeID()] != 6 {
		t.Fatalf("forward dependency simulated incorrectly: %v", p.values)
	}
	Aig_MvSimulateFrame(p, 0, 0)
	for _, e := range []Edge{later, early, po, lo} {
		if p.values[e.NodeID()] != SAIG_UNDEF_VALUE {
			t.Errorf("object %d did not become undefined", e)
		}
	}
	if !reflect.DeepEqual(before, g.Nodes()) {
		t.Error("simulation changed source graph fields")
	}
}

func TestMultivalueOscillatorSearchAgainstCompiledSource(t *testing.T) {
	g := NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	var input, want strings.Builder
	check := func(values []int32) {
		fmt.Fprint(&input, len(values))
		p.states = make([][]int32, len(values))
		hashValues := make([]uint32, len(values))
		for i, v := range values {
			fmt.Fprintf(&input, " %d", v)
			p.states[i] = []int32{v}
			hashValues[i] = uint32(v)
		}
		input.WriteByte('\n')
		fmt.Fprintln(&want, Aig_MvManCheckOscilator(p, 0), Aig_MvSimHash(hashValues, int32(len(values)), 503))
	}
	// Exhaust every binary prefix through length ten and every three-symbol
	// prefix through length eight, with the required initial zero value.
	for _, base := range []int{2, 3} {
		limit := 10
		if base == 3 {
			limit = 8
		}
		count := 1
		for n := 1; n <= limit; n++ {
			for mask := 0; mask < count; mask++ {
				values := make([]int32, n)
				values[0] = 1
				for i, v := 1, mask; i < n; i, v = i+1, v/base {
					values[i] = []int32{1, 0, 2}[v%base]
				}
				check(values)
			}
			count *= base
		}
	}
	// Exercise unsigned hash overflow and wraparound of its sixteen primes.
	random := NewAigRandomSource()
	for _, n := range []int{16, 17, 32, 64} {
		for trial := 0; trial < 16; trial++ {
			values := make([]int32, n)
			values[0] = 1
			for i := 1; i < n; i++ {
				values[i] = int32(random.Uint32() & uint32(SAIG_UNDEF_VALUE))
			}
			check(values)
		}
	}
	got := runAigSourceFixtureInput(t, "multivalue_oscillator_compare.c", input.String())
	if got != want.String() {
		t.Fatal("oscillator candidate search differs from original C")
	}
}

func TestMultivaluePostprocessGroupsActualDefinedHistories(t *testing.T) {
	g := NewGraph()
	for i := 0; i < 4; i++ {
		lo := g.NewInput()
		g.AddOutput(lo)
	}
	g.nRegs = 4
	p := Aig_MvManStart(g, 2)
	p.states = [][]int32{{1, 1, 1, 1}, {0, 0, 1, SAIG_UNDEF_VALUE}, {1, 1, 1, SAIG_UNDEF_VALUE}}
	text := captureMvOutput(t, func() { Aig_MvManPostProcess(p, 1) })
	for _, part := range []string{"never became undef = 3. (Total = 4.)", "FLOP     0 : (  2)", "FLOP     2 : (  1)"} {
		if !strings.Contains(text, part) {
			t.Fatalf("diagnostic missing %q: %s", part, text)
		}
	}
	if strings.Contains(text, "FLOP     3") || strings.Count(text, " # ") != 2 {
		t.Fatalf("undefined history or lasso marker was misclassified: %s", text)
	}
}

func captureMvOutput(t *testing.T, action func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "mv-output")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	old := os.Stdout
	func() {
		defer func() { os.Stdout = old }()
		os.Stdout = f
		action()
	}()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestMultivalueAggressiveSaturationAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "multivalue_saturation_compare.c", "aigMem.c")
	g := NewGraph()
	los := make([]Edge, 6)
	for i := range los {
		los[i] = g.NewInput()
	}
	g.AddOutput(los[5])
	g.AddOutput(los[0].Not())
	for _, e := range los[:5] {
		g.AddOutput(e)
	}
	g.nRegs = 6
	var mapping []Edge
	trace := captureMvOutput(t, func() { mapping = Aig_MvManSimulate(g, 1, 2, 0, 1) })
	var got strings.Builder
	got.WriteString(trace)
	got.WriteString("driver:")
	for _, e := range mapping {
		fmt.Fprintf(&got, " %d", e)
	}
	got.WriteByte('\n')
	if got.String() != out {
		t.Fatalf("Go aggressive saturation:\n%sC aggressive saturation:\n%s", got.String(), out)
	}
	// There are no PIs, so undefined values can arise only from retirement.
	if !strings.Contains(trace, "*") || strings.Count(trace, "\n") < 6 {
		t.Fatalf("fixture missed aggressive retirement: %s", trace)
	}
}
