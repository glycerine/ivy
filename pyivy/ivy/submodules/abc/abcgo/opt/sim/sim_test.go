package sim

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func truthGraph(n int, truth uint64) *aig.Graph {
	g := aig.NewGraph()
	vars := make([]aig.Edge, n)
	for i := range vars {
		vars[i] = g.NewInput()
	}
	var rec func(int, uint64) aig.Edge
	rec = func(level int, f uint64) aig.Edge {
		if level == 0 {
			if f&1 != 0 {
				return g.Const1()
			}
			return g.Const0()
		}
		half := uint(1 << (level - 1))
		return g.Mux(vars[level-1], rec(level-1, f>>half), rec(level-1, f))
	}
	g.AddOutput(rec(n, truth))
	return g
}
func truthTable(t *testing.T, g *aig.Graph) [][]bool {
	t.Helper()
	n := g.InputCount()
	r := make([][]bool, 1<<n)
	for x := range r {
		in := make([]bool, n)
		for i := range in {
			in[i] = x&(1<<i) != 0
		}
		var err error
		r[x], err = g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}
func checkSupportSymmetry(t *testing.T, g *aig.Graph, full bool) {
	t.Helper()
	table := truthTable(t, g)
	support := Sim_ComputeFunSupp(g, false)
	var matrices []Matrix
	if full {
		matrices = Symmetries(g, false)
	}
	for o := range g.Outputs() {
		for a := 0; a < g.InputCount(); a++ {
			dependent := false
			for x := range table {
				dependent = dependent || table[x][o] != table[x^(1<<a)][o]
			}
			if Sim_HasBit(support[o], a) != dependent {
				t.Fatalf("support out%d var%d: %v expected%v", o, a, support, dependent)
			}
			different, model := cofactorDifference(g, o, a, -1)
			if different != dependent {
				t.Fatal("cofactor dependency mismatch")
			}
			if different {
				in := make([]bool, g.InputCount())
				for i := range in {
					in[i] = Sim_HasBit(model, i)
				}
				if !Sim_SolveSuppModelVerify(g, in, a, o) {
					t.Fatal("invalid support SAT witness")
				}
			}
			for b := a + 1; b < g.InputCount(); b++ {
				symmetric := true
				for x := range table {
					y := x
					if (x>>a)&1 != (x>>b)&1 {
						y ^= (1 << a) | (1 << b)
					}
					symmetric = symmetric && table[x][o] == table[y][o]
				}
				different, model := cofactorDifference(g, o, a, b)
				if different == symmetric {
					t.Fatalf("swap%d,%d differs%v symmetric%v", a, b, different, symmetric)
				}
				if different {
					in := 0
					for i := 0; i < g.InputCount(); i++ {
						if Sim_HasBit(model, i) {
							in |= 1 << i
						}
					}
					if table[in^(1<<a)][o] == table[in^(1<<b)][o] {
						t.Fatal("invalid swap SAT witness")
					}
				}
				if full && dependent && Sim_HasBit(support[o], b) && matrices[o].Has(a, b) != symmetric {
					t.Fatalf("workflow symmetry%d,%d wrong", a, b)
				}
			}
		}
	}
}
func TestAllThreeVariableTruthFunctions(t *testing.T) {
	for f := uint64(0); f < 256; f++ {
		checkSupportSymmetry(t, truthGraph(3, f), true)
	}
}
func TestRandomRichGatesAndMultipleOutputs(t *testing.T) {
	r := rand.New(rand.NewSource(19))
	for trial := 0; trial < 40; trial++ {
		g := aig.NewGraph()
		es := []aig.Edge{g.Const0(), g.Const1()}
		for i := 0; i < 4; i++ {
			es = append(es, g.NewInput())
		}
		for i := 0; i < 12; i++ {
			a, b := es[r.Intn(len(es))], es[r.Intn(len(es))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			var e aig.Edge
			if i%3 == 0 {
				e = g.NewXor(a, b)
			} else {
				e = g.AND(a, b)
			}
			es = append(es, e)
		}
		for _, e := range es[len(es)-3:] {
			g.AddOutput(e)
		}
		checkSupportSymmetry(t, g, true)
	}
}
func TestSupportTargetsRequireSAT(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	y := g.NewInput()
	redundant := g.OR(x, g.AND(x, y))
	g.AddOutput(redundant)
	p := Sim_ManStart(g, false)
	Sim_ComputeSuppSetTargets(p)
	Sim_SolveTargetsUsingSat(p, 10)
	if p.SAT != 1 || p.UNSAT != 1 || targetCount(p) != 0 {
		t.Fatalf("SAT%d UNSAT%d targets%d", p.SAT, p.UNSAT, targetCount(p))
	}
	if !Sim_HasBit(p.Functional[0], 0) || Sim_HasBit(p.Functional[0], 1) {
		t.Fatal(p.Functional)
	}
	// A rare minterm prevents finite random simulation from being the proof.
	h := aig.NewGraph()
	root := h.Const1()
	for i := 0; i < 24; i++ {
		root = h.AND(root, h.NewInput())
	}
	h.AddOutput(root)
	supp := Sim_ComputeFunSupp(h, false)
	if Sim_UtilCountOnes(supp[0], 1) != 24 {
		t.Fatal("lost rare dependencies")
	}
}
func TestSequentialTraceAndInitialState(t *testing.T) {
	g := aig.NewGraph()
	x, q := g.NewInput(), g.NewInput()
	out := g.AddOutput(q)
	next := g.AddOutput(g.NewXor(x, q))
	aig.Aig_ManSetRegNum(g, 1)
	model := []bool{true, false, true, true}
	info := Sim_SimulateSeqModel(g, 4, model)
	want := []uint32{0, ^uint32(0), ^uint32(0), 0}
	if !reflect.DeepEqual(info[out.NodeID()], want) {
		t.Fatal(info[out.NodeID()])
	}
	if info[next.NodeID()][3] != ^uint32(0) {
		t.Fatal("wrong final RI")
	}
	one := SequentialSimulation(g, 4, 1, model, []int8{1})
	for f, v := range want {
		if one[out.NodeID()][f] != ^v {
			t.Fatal("initial one not preserved")
		}
	}
	unknown := SequentialSimulation(g, 2, 3, nil, []int8{-1})
	if unknown[q.NodeID()][0] == 0 || unknown[q.NodeID()][0] == ^uint32(0) {
		t.Fatal("unknown state not randomized")
	}
	for w := 0; w < 3; w++ {
		if unknown[q.NodeID()][3+w] != unknown[next.NodeID()][w] {
			t.Fatal("state transfer mismatch")
		}
	}
}
func TestEmptyConstantAndWideNetworks(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	if len(Sim_ComputeFunSupp(g, false)[0]) != 0 || Sim_ComputeTwoVarSymms(g, false) != 0 {
		t.Fatal("constant support")
	}
	info := Sim_SimulateSeqRandom(g, 3, 2)
	for _, w := range info[g.Outputs()[0].NodeID()] {
		if w != ^uint32(0) {
			t.Fatal("constant propagation")
		}
	}
	if len(Sim_SimulateSeqRandom(g, 0, 1)[0]) != 0 {
		t.Fatal("zero frames")
	}
	h := aig.NewGraph()
	var xs []aig.Edge
	for i := 0; i < 70; i++ {
		xs = append(xs, h.NewInput())
	}
	h.AddOutput(h.NewXor(xs[0], xs[69]))
	support := Sim_ComputeFunSupp(h, false)
	if !reflect.DeepEqual(support[0], []uint32{1, 0, 32}) {
		t.Fatal(support)
	}
}
func TestWordUtilitiesAndSwitching(t *testing.T) {
	a, b := []uint32{0x55, 1}, []uint32{0xaa, 3}
	if !reflect.DeepEqual(Sim_UtilInfoDetectDiffs(a, b, 2), []int{0, 1, 2, 3, 4, 5, 6, 7, 33}) {
		t.Fatal("diff bits")
	}
	if !reflect.DeepEqual(Sim_UtilInfoDetectNews(a, b, 2), []int{1, 3, 5, 7, 33}) {
		t.Fatal("new bits")
	}
	if Sim_ComputeSwitching([]uint32{0x55555555}, 1) != 0.5 {
		t.Fatal("switching formula")
	}
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.AddOutput(g.Const1())
	s := Sim_NtkComputeSwitching(g, 3200)
	if s[x.NodeID()] < 0.49 || s[g.Outputs()[1].NodeID()] != 0 {
		t.Fatal(s)
	}
}
func TestParallelAnalysisDoesNotTouchGraph(t *testing.T) {
	g := truthGraph(3, 0x96)
	before := g.Nodes()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Sim_ComputeFunSupp(g, false)
			Symmetries(g, false)
			Sim_SimulateSeqRandom(g, 3, 2)
		}()
	}
	wg.Wait()
	if !reflect.DeepEqual(before, g.Nodes()) {
		t.Fatal("graph scratch mutated")
	}
}

func TestTruncatedFifoNeighborhood(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 2100; i++ {
		g.NewInput()
	}
	p := Sim_ManStart(g, false)
	pat := Sim_ManPatAlloc(p)
	Sim_SetBit(pat.Data, 2099)
	p.Patterns = append(p.Patterns, pat)
	Sim_UtilAssignFromFifo(p)
	if len(p.Patterns) != 0 {
		t.Fatal("FIFO not consumed")
	}
	last := p.Sim0[p.Inputs[2099].NodeID()]
	for _, w := range last {
		if w != ^uint32(0) {
			t.Fatal("truncated flip corrupted row")
		}
	}
	first := p.Sim0[p.Inputs[0].NodeID()]
	if first[0] != 1 {
		t.Fatal("distance-one point missing")
	}
}
func TestMixedImagesAndForwardBuffer(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	buf := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{gate, aig.InvalidEdge}})
	newer := g.NewXor(y, z)
	aig.Aig_ObjConnect(g, buf, newer, aig.InvalidEdge)
	g.AddOutput(g.NewXor(x, buf))
	checkSupportSymmetry(t, g, true)
}
func TestSATClassClosureWithoutRandomRejection(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.NewXor(g.NewXor(a, b), g.AND(c, d)))
	p := Sym_ManStart(g, false)
	pat := make([]uint32, p.Words)
	for Sim_SymmsGetPatternUsingSat(p, pat) {
	}
	Sim_UtilCountPairsAll(p)
	if p.PairsSym[0] != 2 || p.PairsNonSym[0] != 4 || !Sim_UtilMatrsAreDisjoint(p) {
		t.Fatalf("classes sym%v nonsym%v", p.PairsSym, p.PairsNonSym)
	}
}

func TestLightweightManagerSupportStatistics(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	p := Sim_ManStart(g, true)
	if Sim_UtilCountSuppSizes(p, true) != 0 || Sim_UtilCountSuppSizes(p, false) != 0 {
		t.Fatal("lightweight manager reported uncomputed supports")
	}
	Sim_ManStop(p)
	if p.Graph != nil {
		t.Fatal("manager storage retained")
	}
}
