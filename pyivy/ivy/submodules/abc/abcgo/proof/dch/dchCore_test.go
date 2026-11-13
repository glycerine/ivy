package dch

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func associativeSnapshots() ([]*aig.Graph, aig.Edge, aig.Edge) {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, b, c := first.NewInput(), first.NewInput(), first.NewInput()
	x := first.AND(first.AND(a, b), c)
	first.AddOutput(x)
	d, e, f := second.NewInput(), second.NewInput(), second.NewInput()
	y := second.AND(d, second.AND(e, f))
	second.AddOutput(y)
	return []*aig.Graph{first, second}, x, y
}
func checkAllOutputs(t *testing.T, source, result *aig.Graph) {
	t.Helper()
	if source.InputCount() != result.InputCount() || source.OutputCount() != result.OutputCount() {
		t.Fatal("changed interface")
	}
	if err := result.Check(); err != nil {
		t.Fatal(err)
	}
	for pattern := 0; pattern < 1<<source.InputCount(); pattern++ {
		inputs := make([]bool, source.InputCount())
		for i := range inputs {
			inputs[i] = pattern&(1<<i) != 0
		}
		before, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		after, err := result.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal(pattern, before, after, err)
		}
	}
}
func TestSATProvenAssociativeChoices(t *testing.T) {
	variants, _, _ := associativeSnapshots()
	total := Dch_DeriveTotalAig(variants)
	result, stats, err := ComputeChoicesWithStats(total, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, variants[0], result)
	if stats.Proofs == 0 || stats.SATCalls < 1 || aig.Aig_ManChoiceNum(result) == 0 {
		t.Fatal("no SAT-proven choices", stats, aig.Aig_ManChoiceNum(result))
	}
	for _, node := range result.LiveEdges() {
		for next := aig.Aig_ObjEquiv(result, node); next != aig.InvalidEdge; next = aig.Aig_ObjEquiv(result, next) {
			for pattern := 0; pattern < 8; pattern++ {
				input := []bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0}
				order, _ := graphOrder(result)
				values := simulateAssignment(result, order, input)
				a, b := values[node.NodeID()], values[next.NodeID()]
				if (a != b) != (aig.Aig_ObjPhase(result, node) != aig.Aig_ObjPhase(result, next)) {
					t.Fatal("unsound choice")
				}
			}
		}
	}
}

func TestConcurrentIndependentChoiceNetworks(t *testing.T) {
	const workers = 12
	errors := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			variants, _, _ := associativeSnapshots()
			total, err := DeriveTotalAig(variants)
			if err != nil {
				errors <- err
				return
			}
			result, err := ComputeChoices(total, nil)
			if err != nil {
				errors <- err
				return
			}
			if err := CheckChoices(result, false); err != nil {
				errors <- err
				return
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				got, err := result.Evaluate(inputs)
				if err != nil {
					errors <- err
					return
				}
				if len(got) != 1 || got[0] != (inputs[0] && inputs[1] && inputs[2]) {
					errors <- fmt.Errorf("input %v produced %v", inputs, got)
					return
				}
			}
			errors <- nil
		}()
	}
	for i := 0; i < workers; i++ {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
}
func TestRareMintermRequiresSATCounterexample(t *testing.T) {
	g := aig.NewGraph()
	root := g.Const1()
	for i := 0; i < 30; i++ {
		root = g.AND(root, g.NewInput())
	}
	g.AddOutput(root)
	result, stats, err := ComputeChoicesWithStats(g, &Params{Words: 1, PolarFlip: true, SimulateTFO: true, ConflictLimit: 1000, SatVarMax: 5000, CallsRecycle: 100})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Counterexamples == 0 {
		t.Fatal("rare minterm was not challenged by SAT", stats)
	}
	for _, value := range []bool{false, true} {
		in := make([]bool, 30)
		for i := range in {
			in[i] = value
		}
		got, err := result.Evaluate(in)
		if err != nil || got[0] != value {
			t.Fatal(value, got, err)
		}
	}
}
func TestComplementAndConstantRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	and := g.AND(x, y)
	notAnd := g.OR(x.Not(), y.Not())
	inverse := g.AND(g.OR(notAnd, z), g.OR(notAnd, z.Not()))
	zero := g.AND(g.AND(x, y), g.AND(x, y.Not()))
	g.AddOutput(and)
	g.AddOutput(inverse)
	g.AddOutput(zero)
	result, stats, err := ComputeChoicesWithStats(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, g, result)
	if stats.Proofs == 0 {
		t.Fatal("no constant/complement proofs", stats)
	}
	aig.Aig_ManReprStop(g)
	if err := ComputeEquivalences(g, nil); err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ObjRepr(g, inverse) == aig.InvalidEdge || aig.Aig_ObjRepr(g, zero) != g.Const1() {
		t.Fatal("missing representative", aig.Aig_ObjRepr(g, inverse), aig.Aig_ObjRepr(g, zero))
	}
}
func TestPolarityTFOAndRecyclingSettings(t *testing.T) {
	for _, polar := range []bool{false, true} {
		for _, tfo := range []bool{false, true} {
			variants, _, _ := associativeSnapshots()
			total := Dch_DeriveTotalAig(variants)
			p := DefaultParams()
			p.PolarFlip = polar
			p.SimulateTFO = tfo
			p.SatVarMax = 1
			p.CallsRecycle = 0
			result, stats, err := ComputeChoicesWithStats(total, &p)
			if err != nil {
				t.Fatal(err)
			}
			checkAllOutputs(t, variants[0], result)
			if stats.Recycles == 0 || stats.Proofs == 0 {
				t.Fatal(stats)
			}
		}
	}
}
func TestEngineAndDefaultParameters(t *testing.T) {
	defaults := DefaultParams()
	var params Params
	Dch_ManSetDefaultParams(&params)
	if params != defaults || params.Words != 8 || params.ConflictLimit != 1000 || params.SatVarMax != 5000 {
		t.Fatal(params)
	}
	variants, _, _ := associativeSnapshots()
	result, err := (Engine{}).ComputeChoices(variants)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, variants[0], result)
	if _, err := ComputeChoices(nil, nil); err == nil {
		t.Fatal("accepted nil graph")
	}
	params.Words = 0
	if _, err := ComputeChoices(variants[0], &params); err == nil {
		t.Fatal("accepted zero simulation words")
	}
}

func TestConflictLimitUnknownNeverCreatesChoice(t *testing.T) {
	left, right := aig.NewGraph(), aig.NewGraph()
	a, b := make([]aig.Edge, 8), make([]aig.Edge, 8)
	for i := range a {
		a[i] = left.NewInput()
		b[i] = right.NewInput()
	}
	x := a[0]
	for _, next := range a[1:] {
		x = left.XOR(x, next)
	}
	left.AddOutput(x)
	var parity func([]aig.Edge) aig.Edge
	parity = func(nodes []aig.Edge) aig.Edge {
		if len(nodes) == 1 {
			return nodes[0]
		}
		mid := len(nodes) / 2
		return right.XOR(parity(nodes[:mid]), parity(nodes[mid:]))
	}
	right.AddOutput(parity(b))
	total := Dch_DeriveTotalAig([]*aig.Graph{left, right})
	params := DefaultParams()
	params.ConflictLimit = 1
	result, stats, err := ComputeChoicesWithStats(total, &params)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, left, result)
	if stats.Unknown == 0 {
		t.Fatal("expected bounded proof to remain unknown", stats)
	}
}
func redundantSupportVariants() []*aig.Graph {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, _, c := first.NewInput(), first.NewInput(), first.NewInput()
	first.AddOutput(first.AND(a, c))
	x, y, z := second.NewInput(), second.NewInput(), second.NewInput()
	second.AddOutput(second.OR(second.AND(second.AND(x, y), z), second.AND(second.AND(x, y.Not()), z)))
	return []*aig.Graph{first, second}
}
func TestSkipRedundantSupportAndChoiceCycles(t *testing.T) {
	var counts []int32
	for _, skip := range []bool{false, true} {
		variants := redundantSupportVariants()
		total := Dch_DeriveTotalAig(variants)
		params := DefaultParams()
		params.SkipRedundantSupport = skip
		result, stats, err := ComputeChoicesWithStats(total, &params)
		if err != nil {
			t.Fatal(err)
		}
		checkAllOutputs(t, variants[0], result)
		counts = append(counts, aig.Aig_ManChoiceNum(result))
		if stats.Proofs == 0 {
			t.Fatal("no proven equivalence")
		}
	}
	if counts[0] == 0 || counts[1] != 0 {
		t.Fatal("redundant support filter ineffective", counts)
	}
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	head := g.AND(a, b)
	member := g.AND(head, g.OR(a, c))
	g.AddOutput(member)
	result, stats, err := ComputeChoicesWithStats(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, g, result)
	if stats.Proofs == 0 || aig.Aig_ManChoiceNum(result) != 0 {
		t.Fatal("cyclic alternate admitted", stats)
	}
}
func TestClassCounterexampleHistoryRefinement(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetPhase(g)
	for _, tfo := range []bool{false, true} {
		params := DefaultParams()
		params.SimulateTFO = tfo
		m, err := newManager(g, &params)
		if err != nil {
			t.Fatal(err)
		}
		m.classes = &Classes{groups: [][]aig.Edge{{a, b, c}}}
		m.classes.index()
		m.refineCounterexample([]bool{true, false, false})
		if m.classes.representative(b) != aig.InvalidEdge || m.classes.representative(c) != b {
			t.Fatal("first partition wrong", m.classes.groups)
		}
		m.refineCounterexample([]bool{true, false, true})
		if m.classes.representative(c) != aig.InvalidEdge {
			t.Fatal("second pattern ignored", m.classes.groups)
		}
		m.refineCounterexample([]bool{true, false, true})
		if len(m.classes.groups) != 3 {
			t.Fatal("stable pattern merged classes", m.classes.groups)
		}
	}
}

func TestSequentialInterfaceMetadataAndTrailingRegisters(t *testing.T) {
	variants := make([]*aig.Graph, 2)
	for i := range variants {
		g := aig.NewGraph()
		pi, state := g.NewInput(), g.NewInput()
		g.AddOutput(g.XOR(pi, state))
		g.AddOutput(g.AND(pi, state))
		aig.Aig_ManSetRegNum(g, 1)
		g.Name = "machine"
		g.Spec = "source.blif"
		g.FlopNumbers = []int32{17}
		g.ClockDomains = [][]int32{{0}}
		variants[i] = g
	}
	total, err := DeriveTotalAig(variants)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ComputeChoices(total, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, variants[0], result)
	if aig.Aig_ManRegNum(total) != 1 || aig.Aig_ManRegNum(result) != 1 || result.Name != "machine" || result.Spec != "source.blif" || result.FlopNumbers[0] != 17 {
		t.Fatal("lost register metadata")
	}
	result.FlopNumbers[0] = 99
	result.ClockDomains[0][0] = 99
	if variants[0].FlopNumbers[0] != 17 || variants[0].ClockDomains[0][0] != 0 {
		t.Fatal("shared mutable metadata")
	}
	aig.Aig_ManSetRegNum(variants[1], 0)
	if _, err := DeriveTotalAig(variants); err == nil {
		t.Fatal("accepted different register interfaces")
	}
}
func TestSolverActuallyRecyclesBetweenProofs(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(a, b), g.AND(a, b.Not())))
	g.AddOutput(g.AND(g.AND(a, c), g.AND(a, c.Not())))
	p := DefaultParams()
	p.SatVarMax = 1
	p.CallsRecycle = 0
	result, stats, err := ComputeChoicesWithStats(g, &p)
	if err != nil {
		t.Fatal(err)
	}
	checkAllOutputs(t, g, result)
	if stats.Proofs < 2 || stats.Recycles < 2 {
		t.Fatal("solver was not recycled across actual proofs", stats)
	}
}

func TestOriginalDCHOracleSnapshotsAndSupportFiltering(t *testing.T) {
	data, err := os.ReadFile("test_vectors/dch_oracle.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatal("incomplete C oracle")
	}
	for _, line := range lines {
		var which, totalNodes, resultNodes, choiceCount, truth int
		if _, err := fmt.Sscan(line, &which, &totalNodes, &resultNodes, &choiceCount, &truth); err != nil {
			t.Fatal(err)
		}
		var variants []*aig.Graph
		if which == 0 {
			variants, _, _ = associativeSnapshots()
		} else {
			variants = redundantSupportVariants()
		}
		total := Dch_DeriveTotalAig(variants)
		if int(aig.Aig_ManNodeNum(total)) != totalNodes {
			t.Fatal("total nodes disagree with original C", line, aig.Aig_ManNodeNum(total))
		}
		params := DefaultParams()
		params.SkipRedundantSupport = which == 2
		result, err := ComputeChoices(total, &params)
		if err != nil {
			t.Fatal(err)
		}
		if int(aig.Aig_ManNodeNum(result)) != resultNodes || int(aig.Aig_ManChoiceNum(result)) != choiceCount {
			t.Fatalf("case %d: nodes/choices %d/%d, C %d/%d", which, aig.Aig_ManNodeNum(result), aig.Aig_ManChoiceNum(result), resultNodes, choiceCount)
		}
		gotTruth := 0
		for pattern := 0; pattern < 8; pattern++ {
			in := []bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0}
			out, err := result.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			if out[0] {
				gotTruth |= 1 << pattern
			}
		}
		if gotTruth != truth {
			t.Fatal("truth disagrees with C", which, gotTruth, truth)
		}
	}
}

type timingSnapshot struct{ calls *int }

func (s timingSnapshot) CloneTiming() any { *s.calls++; return s }
func TestChoiceMetadataClonedOnceAndBadRepresentativesRejected(t *testing.T) {
	variants, _, _ := associativeSnapshots()
	total := Dch_DeriveTotalAig(variants)
	calls := 0
	total.Timing = timingSnapshot{&calls}
	if _, err := ComputeChoices(total, nil); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("choice metadata cloned redundantly", calls)
	}
	if _, err := DeriveChoiceAig(nil, false); err == nil {
		t.Fatal("accepted nil choice source")
	}
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	g.AddOutput(n)
	aig.Aig_ObjSetRepr(g, n, aig.Edge(9999))
	if _, err := DeriveChoiceAig(g, false); err == nil {
		t.Fatal("accepted foreign representative")
	}
}

func TestRandomVariantsEveryChoiceIsSemanticallyValid(t *testing.T) {
	random := rand.New(rand.NewSource(2157))
	for run := 0; run < 25; run++ {
		first, second := aig.NewGraph(), aig.NewGraph()
		a, b := make([]aig.Edge, 5), make([]aig.Edge, 5)
		for i := range a {
			a[i] = first.NewInput()
			b[i] = second.NewInput()
		}
		for i := 0; i < 18; i++ {
			x, y := random.Intn(len(a)), random.Intn(len(a))
			negX, negY := random.Intn(2) == 0, random.Intn(2) == 0
			ax, ay, bx, by := phaseEdge(a[x], negX), phaseEdge(a[y], negY), phaseEdge(b[x], negX), phaseEdge(b[y], negY)
			if random.Intn(3) == 0 {
				a = append(a, first.XOR(ax, ay))
				b = append(b, second.XOR(bx, by))
			} else {
				t := b[random.Intn(5)]
				a = append(a, first.AND(ax, ay))
				b = append(b, second.AND(bx, second.AND(second.OR(by, t), second.OR(by, t.Not()))))
			}
		}
		for i := len(a) - 3; i < len(a); i++ {
			first.AddOutput(a[i])
			second.AddOutput(b[i])
		}
		total := Dch_DeriveTotalAig([]*aig.Graph{first, second})
		result, err := ComputeChoices(total, nil)
		if err != nil {
			t.Fatal(run, err)
		}
		checkAllOutputs(t, first, result)
		if err := CheckChoices(result, false); err != nil {
			t.Fatal(run, err)
		}
		sourceOrder, err := graphOrder(total)
		if err != nil {
			t.Fatal(err)
		}
		for pattern := 0; pattern < 32; pattern++ {
			input := make([]bool, 5)
			for i := range input {
				input[i] = pattern&(1<<i) != 0
			}
			values := simulateAssignment(total, sourceOrder, input)
			for _, node := range total.LiveEdges() {
				representative := aig.Aig_ObjRepr(total, node)
				if representative == aig.InvalidEdge {
					continue
				}
				if (values[node.NodeID()] != values[representative.NodeID()]) != (aig.Aig_ObjPhase(total, node) != aig.Aig_ObjPhase(total, representative)) {
					t.Fatal("unsound published proof", run, pattern, node, representative)
				}
			}
		}
		order, err := graphOrder(result)
		if err != nil {
			t.Fatal(err)
		}
		for pattern := 0; pattern < 32; pattern++ {
			inputs := make([]bool, 5)
			for i := range inputs {
				inputs[i] = pattern&(1<<i) != 0
			}
			values := simulateAssignment(result, order, inputs)
			for _, node := range result.LiveEdges() {
				next := aig.Aig_ObjEquiv(result, node)
				if next == aig.InvalidEdge {
					continue
				}
				if (values[node.NodeID()] != values[next.NodeID()]) != (aig.Aig_ObjPhase(result, node) != aig.Aig_ObjPhase(result, next)) {
					t.Fatal("unsound random choice", run, pattern, node, next)
				}
			}
		}
	}
}
func TestChoiceCheckerRejectsCombinedCycles(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	head := g.AND(a, b)
	member := g.AND(head, c)
	g.AddOutput(head)
	aig.Aig_ObjSetEquiv(g, head, member)
	if err := CheckChoices(g, false); err == nil {
		t.Fatal("combined logic/choice cycle accepted")
	}
}

func TestNonTopologicalNodeIDsAreRejected(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	early := g.AND(a, b)
	later := g.AND(b, c)
	aig.Aig_ObjDisconnect(g, early)
	aig.Aig_ObjConnect(g, early, a, later)
	alternative := g.AND(g.AND(a, b), c)
	g.AddOutput(early)
	g.AddOutput(alternative.Not())
	if result, _, err := ComputeChoicesWithStats(g, nil); err == nil || result != nil {
		t.Fatalf("non-topological graph accepted: result=%v error=%v", result, err)
	}
}

func TestIncrementalTFOSimulationMatchesFullValues(t *testing.T) {
	random := rand.New(rand.NewSource(921))
	g := aig.NewGraph()
	nodes := make([]aig.Edge, 5)
	for i := range nodes {
		nodes[i] = g.NewInput()
	}
	for i := 0; i < 80; i++ {
		a, b := nodes[random.Intn(len(nodes))], nodes[random.Intn(len(nodes))]
		if random.Intn(2) == 0 {
			a = a.Not()
		}
		if random.Intn(3) == 0 {
			nodes = append(nodes, g.XOR(a, b))
		} else {
			nodes = append(nodes, g.AND(a, b))
		}
	}
	g.AddOutput(nodes[len(nodes)-1])
	m, err := newManager(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 100; trial++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = random.Intn(2) != 0
		}
		m.simulateTFO(inputs)
		full := simulateAssignment(g, m.order, inputs)
		if !reflect.DeepEqual(m.lastValues, full) {
			t.Fatal("incremental simulation differs", trial)
		}
	}
}

func TestConstantsEmptyAndDirectInputInterfaces(t *testing.T) {
	for _, which := range []int{0, 1, 2} {
		g := aig.NewGraph()
		if which == 1 {
			g.AddOutput(g.Const1())
			g.AddOutput(g.Const0())
		}
		if which == 2 {
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(a)
			g.AddOutput(b.Not())
		}
		result, stats, err := ComputeChoicesWithStats(g, nil)
		if err != nil {
			t.Fatal(err)
		}
		checkAllOutputs(t, g, result)
		if stats.SATCalls != 0 || aig.Aig_ManChoiceNum(result) != 0 {
			t.Fatal("invented non-gate proofs", stats)
		}
	}
}

func TestSimulationDimensionOverflowIsError(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	p := DefaultParams()
	p.Words = int(^uint(0) >> 1)
	if _, err := ComputeChoices(g, &p); err == nil {
		t.Fatal("simulation dimension overflow accepted")
	}
}
