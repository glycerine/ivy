package sbd

import (
	"math/bits"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func graphTruth(t *testing.T, graph *aig.Graph) [][]bool {
	t.Helper()
	result := make([][]bool, 1<<graph.InputCount())
	for assignment := range result {
		inputs := make([]bool, graph.InputCount())
		for k := range inputs {
			inputs[k] = assignment>>k&1 != 0
		}
		var err error
		result[assignment], err = graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}
func checkSameTruth(t *testing.T, before [][]bool, after *aig.Graph) {
	t.Helper()
	actual := graphTruth(t, after)
	if len(before) != len(actual) {
		t.Fatal("input count changed")
	}
	for i := range before {
		if !slices.Equal(before[i], actual[i]) {
			t.Fatalf("assignment%x before%v after%v", i, before[i], actual[i])
		}
	}
}
func TestOptimizerUsesInternalDontCares(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	pivot := graph.AND(x, c)
	graph.AddOutput(graph.AND(pivot, a.Not()))
	network := NewNetwork(graph)
	before := graphTruth(t, graph)
	var parameters Parameters
	Sbd_ParSetDefault(&parameters)
	parameters.LUTSize, parameters.LUTCount, parameters.CutSize = 2, 2, 4
	result := Sbd_NtkPerform(network, &parameters)
	checkSameTruth(t, before, result.Graph)
	if result.Graph.Node(result.Graph.Node(result.Graph.Outputs()[0]).Fanins[0]).Kind != aig.Constant {
		t.Fatal("unobservable nonconstant pivot was not eliminated")
	}
}
func TestOptimizerRandomEquivalenceAndMapping(t *testing.T) {
	for _, options := range []struct{ mapping, more, divisors, path bool }{{true, false, false, false}, {true, true, true, false}, {false, false, true, false}, {true, false, true, true}} {
		for seedStart := uint32(1); seedStart <= 6; seedStart++ {
			graph := aig.NewGraph()
			var edges []aig.Edge
			for i := 0; i < 5; i++ {
				edges = append(edges, graph.NewInput())
			}
			seed := seedStart
			for node := 0; node < 25; node++ {
				seed = seed*1664525 + 1013904223
				a := edges[int(seed)%len(edges)]
				seed = seed*1664525 + 1013904223
				b := edges[int(seed)%len(edges)]
				if seed&1 != 0 {
					a = a.Not()
				}
				if seed&2 != 0 {
					b = b.Not()
				}
				edges = append(edges, graph.AND(a, b))
			}
			graph.AddOutput(edges[len(edges)-1])
			graph.AddOutput(edges[len(edges)-3].Not())
			network := NewNetwork(graph)
			before := graphTruth(t, graph)
			var parameters Parameters
			Sbd_ParSetDefault(&parameters)
			parameters.LUTSize, parameters.LUTCount, parameters.CutSize = 3, 2, 5
			parameters.Mapping, parameters.MoreCuts, parameters.FindDivisors, parameters.CriticalPathOnly = options.mapping, options.more, options.divisors, options.path
			t.Logf("options%+v seed%d", options, seedStart)
			result := Sbd_NtkPerform(network, &parameters)
			checkSameTruth(t, before, result.Graph)
			if err := result.Graph.Check(); err != nil {
				t.Fatal(err)
			}
			if options.mapping {
				for root, cut := range result.LUTs {
					if len(cut.Leaves) > int(parameters.LUTSize) || len(cut.Truth) == 0 {
						t.Fatalf("invalid result mapping root%d", root)
					}
				}
			} else if len(result.LUTs) != 0 {
				t.Fatal("mapping emitted when disabled")
			}
		}
	}
}
func TestOptimizerPreservesTimingBoxLinks(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	boxInput := graph.AddOutput(graph.AND(a, b))
	boxOutput := graph.NewInput()
	graph.AddOutput(graph.AND(boxOutput, c))
	network := NewNetwork(graph)
	network.SetBoxOutput(int32(boxOutput.NodeID()), []int32{int32(boxInput.NodeID())}, []bool{true})
	before := graphTruth(t, graph)
	var parameters Parameters
	Sbd_ParSetDefault(&parameters)
	result := Sbd_NtkPerform(network, &parameters)
	checkSameTruth(t, before, result.Graph)
	if len(result.BoxInputs) != 1 || len(result.BoxLevelInputs) != 1 {
		t.Fatal("timing boxes lost")
	}
	for ci, inputs := range result.BoxInputs {
		if result.Graph.Node(aig.Edge(ci)<<1).Kind != aig.Input || len(inputs) != 1 || result.Graph.Node(aig.Edge(inputs[0])<<1).Kind != aig.Output {
			t.Fatal("timing endpoints not remapped")
		}
	}
}
func TestCoverDominanceAndMinimumCandidate(t *testing.T) {
	var rows [64]uint64
	count := Sbd_ManAddCube1(64, rows[:], 0, 7)
	count = Sbd_ManAddCube1(64, rows[:], count, 3)
	count = Sbd_ManAddCube1(64, rows[:], count, 15)
	if count != 1 || rows[0] != 3 {
		t.Fatal("cover row dominance")
	}
	p := &Sbd_Man_t{pPars: &Parameters{LUTSize: 4}, vDivSet: sbdVector()}
	for encoded := uint32(0); encoded < 65536; encoded++ {
		cover := []uint64{uint64(encoded & 15), uint64(encoded >> 4 & 15), uint64(encoded >> 8 & 15), uint64(encoded >> 12 & 15), 15}
		minimum := 5
		for subset := uint32(1); subset < 16; subset++ {
			var union uint64
			for i := 0; i < 4; i++ {
				if subset>>i&1 != 0 {
					union |= cover[i]
				}
			}
			if union == 15 {
				minimum = min(minimum, bits.OnesCount32(subset))
			}
		}
		found := Sbd_ManFindCandsSimple(p, cover, 4)
		if (found != 0) != (minimum < 5) {
			t.Fatal("cover feasibility")
		}
		if found != 0 {
			var union uint64
			for _, index := range vec.Vec_IntArray(p.vDivSet) {
				union |= cover[index]
			}
			if union != 15 || vec.Vec_IntSize(p.vDivSet) != minimum {
				t.Fatal("nonminimal cover candidate")
			}
		}
	}
}

func TestTwoLutResynthesisImprovesExpandedParityCone(t *testing.T) {
	graph := aig.NewGraph()
	inputs := make([]aig.Edge, 6)
	for i := range inputs {
		inputs[i] = graph.NewInput()
	}
	sum := graph.Const0()
	for assignment := 0; assignment < 16; assignment++ {
		if bits.OnesCount32(uint32(assignment))%2 == 0 {
			continue
		}
		cube := graph.Const1()
		for k := 0; k < 4; k++ {
			literal := inputs[k]
			if assignment>>k&1 == 0 {
				literal = literal.Not()
			}
			cube = graph.AND(cube, literal)
		}
		cube = graph.AND(cube, inputs[4])
		sum = graph.OR(sum, cube)
	}
	output := graph.OR(sum, inputs[5])
	graph.AddOutput(output)
	network := NewNetwork(graph)
	before := graphTruth(t, graph)
	var parameters Parameters
	Sbd_ParSetDefault(&parameters)
	parameters.MoreCuts = true
	p := Sbd_ManStart(network, &parameters)
	defer Sbd_ManStop(p)
	for _, edge := range graph.LiveEdges() {
		Sbd_StoRefObj(p.pSto, int32(edge.NodeID()), -1)
	}
	Sbd_StoComputeCutsConst0(p.pSto, 0)
	for _, edge := range graph.Inputs() {
		Sbd_StoComputeCutsCi(p.pSto, int32(edge.NodeID()), 0, 0)
	}
	for _, id := range sbdCoreOrder(network, false) {
		if network.object(id).Kind == aig.And {
			delay := Sbd_StoComputeCutsNode(p.pSto, id)
			vec.Vec_IntWriteEntry(p.vLutLevs, int(id), delay)
			Sbd_StoSaveBestDelayCut(p.pSto, id, Sbd_ObjCut(p, id))
		}
	}
	pivot := int32(output.NodeID())
	old := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	if old < 3 {
		t.Fatalf("expanded cone unexpectedly maps in%d levels", old)
	}
	Sbd_NtkPerformOne(p, pivot)
	if vec.Vec_IntEntry(p.vMirrors, int(pivot)) < 0 {
		t.Fatal("two-LUT factorization was not found")
	}
	if level := vec.Vec_IntEntry(p.vLutLevs, int(pivot)); level >= old || level > 2 {
		t.Fatalf("delay%d -> %d", old, level)
	}
	result := Sbd_ManDerive(p, network, p.vMirrors)
	checkSameTruth(t, before, result.Graph)
	if p.nLuts[2] == 0 {
		t.Fatalf("two-LUT template was not exercised: %v", p.nLuts)
	}
}

func TestImplementationRejectsStructuralSelfMirror(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	root := graph.AND(a, b)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	var parameters Parameters
	Sbd_ParSetDefault(&parameters)
	p := Sbd_ManStart(network, &parameters)
	defer Sbd_ManStop(p)
	for _, edge := range graph.LiveEdges() {
		Sbd_StoRefObj(p.pSto, int32(edge.NodeID()), -1)
	}
	Sbd_StoComputeCutsConst0(p.pSto, 0)
	for _, edge := range graph.Inputs() {
		Sbd_StoComputeCutsCi(p.pSto, int32(edge.NodeID()), 0, 0)
	}
	Sbd_StoComputeCutsNode(p.pSto, int32(root.NodeID()))
	vec.Vec_IntWriteEntry(p.vLutLevs, root.NodeID(), 3)
	p.vWinObjs = sbdVector(0, int32(a.NodeID()), int32(b.NodeID()), int32(root.NodeID()))
	p.vDivSet = sbdVector(1, 2)
	structure := NewLUTStructure([]int32{0, 1})
	structure.Res = 0x8888888888888888
	Sbd_ManImplement2(p, int32(root.NodeID()), 1, []Sbd_Str_t{structure})
	if vec.Vec_IntEntry(p.vMirrors, root.NodeID()) != -1 {
		t.Fatal("structural hashing created a self mirror")
	}
}

func prepareRootOptimization(t *testing.T, graph *aig.Graph) *Sbd_Man_t {
	t.Helper()
	network := NewNetwork(graph)
	parameters := new(Parameters)
	Sbd_ParSetDefault(parameters)
	parameters.MoreCuts = true
	p := Sbd_ManStart(network, parameters)
	t.Cleanup(func() { Sbd_ManStop(p) })
	for _, edge := range graph.LiveEdges() {
		Sbd_StoRefObj(p.pSto, int32(edge.NodeID()), -1)
	}
	Sbd_StoComputeCutsConst0(p.pSto, 0)
	for _, edge := range graph.Inputs() {
		Sbd_StoComputeCutsCi(p.pSto, int32(edge.NodeID()), 0, 0)
	}
	for _, id := range sbdCoreOrder(network, false) {
		if network.object(id).Kind == aig.And {
			delay := Sbd_StoComputeCutsNode(p.pSto, id)
			vec.Vec_IntWriteEntry(p.vLutLevs, int(id), delay)
			Sbd_StoSaveBestDelayCut(p.pSto, id, Sbd_ObjCut(p, id))
		}
	}
	return p
}
func TestThreeParallelLutResynthesis(t *testing.T) {
	graph := aig.NewGraph()
	inputs := make([]aig.Edge, 8)
	for i := range inputs {
		inputs[i] = graph.NewInput()
	}
	sum := graph.Const0()
	for assignment := 0; assignment < 256; assignment++ {
		if bits.OnesCount32(uint32(assignment&15))%2 == 0 || bits.OnesCount32(uint32(assignment>>4))%2 == 0 {
			continue
		}
		cube := graph.Const1()
		for k, literal := range inputs {
			if assignment>>k&1 == 0 {
				literal = literal.Not()
			}
			cube = graph.AND(cube, literal)
		}
		sum = graph.OR(sum, cube)
	}
	graph.AddOutput(sum)
	before := graphTruth(t, graph)
	p := prepareRootOptimization(t, graph)
	pivot := int32(sum.NodeID())
	old := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	Sbd_NtkPerformOne(p, pivot)
	if vec.Vec_IntEntry(p.vMirrors, int(pivot)) < 0 || p.nLuts[3] == 0 {
		t.Fatalf("parallel template not found old%d stats%v", old, p.nLuts)
	}
	if level := vec.Vec_IntEntry(p.vLutLevs, int(pivot)); level > 2 || level >= old {
		t.Fatalf("parallel levels%d ->%d", old, level)
	}
	checkSameTruth(t, before, Sbd_ManDerive(p, p.pGia, p.vMirrors).Graph)
}
func TestThreeCascadeLutResynthesis(t *testing.T) {
	graph := aig.NewGraph()
	inputs := make([]aig.Edge, 10)
	for i := range inputs {
		inputs[i] = graph.NewInput()
	}
	sum := graph.Const0()
	for assignment := 0; assignment < 16; assignment++ {
		if bits.OnesCount32(uint32(assignment))%2 == 0 {
			continue
		}
		cube := graph.Const1()
		for k := 0; k < 4; k++ {
			literal := inputs[k]
			if assignment>>k&1 == 0 {
				literal = literal.Not()
			}
			cube = graph.AND(cube, literal)
		}
		for _, k := range []int{7, 8, 9} {
			cube = graph.AND(cube, inputs[k])
		}
		sum = graph.OR(sum, cube)
	}
	for _, k := range []int{4, 5, 6} {
		sum = graph.OR(sum, inputs[k])
	}
	graph.AddOutput(sum)
	before := graphTruth(t, graph)
	p := prepareRootOptimization(t, graph)
	pivot := int32(sum.NodeID())
	old := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	if Sbd_ManWindow(p, pivot) == 0 || Sbd_ManCheckConst(p, pivot) != -1 {
		t.Fatal("could not build cascade care window")
	}
	leaves := make([]int32, 10)
	for i, edge := range inputs {
		leaves[i] = int32(edge.NodeID())
	}
	free := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	var count int32
	structures := make([]Sbd_Str_t, 10)
	if Sbd_ManExploreCut(p, pivot, 10, leaves, &count, structures, &free) == 0 {
		t.Fatal("wide-cut cascade synthesis failed")
	}
	Sbd_ManImplement2(p, pivot, count, structures)
	if vec.Vec_IntEntry(p.vMirrors, int(pivot)) < 0 || p.nLuts[4] == 0 {
		t.Fatalf("cascade template not found old%d stats%v", old, p.nLuts)
	}
	if level := vec.Vec_IntEntry(p.vLutLevs, int(pivot)); level > 3 || level >= old {
		t.Fatalf("cascade levels%d ->%d", old, level)
	}
	checkSameTruth(t, before, Sbd_ManDerive(p, p.pGia, p.vMirrors).Graph)
}

func TestOptimizerFiveAndSixInputLutTargets(t *testing.T) {
	for _, size := range []int32{5, 6} {
		graph := aig.NewGraph()
		root := graph.Const1()
		for i := 0; i < 8; i++ {
			root = graph.AND(root, graph.NewInput())
		}
		graph.AddOutput(root)
		network := NewNetwork(graph)
		before := graphTruth(t, graph)
		var parameters Parameters
		Sbd_ParSetDefault(&parameters)
		parameters.LUTSize, parameters.CutSize, parameters.LUTCount = size, 10, 2
		parameters.MoreCuts = true
		parameters.FindDivisors = true
		result := Sbd_NtkPerform(network, &parameters)
		checkSameTruth(t, before, result.Graph)
		for _, cut := range result.LUTs {
			if len(cut.Leaves) > int(size) {
				t.Fatal("oversized result cut")
			}
		}
	}
}

func TestOptimizerKeepsObservabilityAcrossBuffers(t *testing.T) {
	for _, critical := range []bool{false, true} {
		graph := aig.NewGraph()
		a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
		pivot := graph.AND(graph.AND(a, b), c)
		buffer := aig.Aig_ObjCreate(graph, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{pivot}})
		graph.AddOutput(graph.AND(pivot, a.Not()))
		graph.AddOutput(buffer)
		network := NewNetwork(graph)
		before := graphTruth(t, graph)
		var parameters Parameters
		Sbd_ParSetDefault(&parameters)
		parameters.LUTSize, parameters.LUTCount, parameters.CutSize = 2, 2, 4
		parameters.CriticalPathOnly = critical
		result := Sbd_NtkPerform(network, &parameters)
		checkSameTruth(t, before, result.Graph)
	}
}

func TestHeuristicCoverCandidatesAreFeasible(t *testing.T) {
	p := &Sbd_Man_t{pPars: &Parameters{LUTSize: 4}, vDivSet: sbdVector()}
	seed := uint64(3)
	for count := int32(8); count <= 20; count++ {
		for iteration := 0; iteration < 200; iteration++ {
			cover := make([]uint64, count+1)
			for i := range cover[:count] {
				seed = seed*6364136223846793005 + 1
				cover[i] = seed >> 32 & 255
			}
			cover[count] = 255
			if Sbd_ManFindCands(p, cover, count) != 0 {
				var union uint64
				for _, column := range vec.Vec_IntArray(p.vDivSet) {
					union |= cover[column]
				}
				if union != 255 || vec.Vec_IntSize(p.vDivSet) > 4 {
					t.Fatal("infeasible heuristic cover")
				}
			}
		}
	}
}
