package ivy

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func ivySatFixture(inputs int) (*Ivy_Man_t, *Ivy_FraigMan_t) {
	source := Ivy_ManStart()
	for i := 0; i < inputs; i++ {
		Ivy_ObjCreatePi(source)
	}
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.NBTLimitNode = 0
	params.NBTLimitMiter = 0
	return source, Ivy_FraigStartSimple(source, &params)
}
func TestFraigSatEquivalenceAndCounterexample(t *testing.T) {
	_, p := ivySatFixture(3)
	graph := p.pManFraig
	a, b, c := graph.vPis[0], graph.vPis[1], graph.vPis[2]
	left, right := Ivy_And(graph, Ivy_And(graph, a, b), c), Ivy_And(graph, a, Ivy_And(graph, b, c))
	if Ivy_FraigNodesAreEquiv(p, left, right) != 1 {
		t.Fatal("associative AND equivalence not proved")
	}
	different := Ivy_And(graph, a, c)
	if Ivy_FraigNodesAreEquiv(p, left, different) != 0 {
		t.Fatal("different functions considered equivalent")
	}
	model := Ivy_FraigCreateModel(p)
	assignment := 0
	for i, value := range model {
		if value != 0 {
			assignment |= 1 << i
		}
	}
	Ivy_ObjCreatePo(graph, left)
	Ivy_ObjCreatePo(graph, different)
	values := ivyBalanceEvaluate(graph, assignment)
	if values[0] == values[1] {
		t.Fatal("SAT model does not distinguish candidates")
	}
	if Ivy_FraigNodesAreEquivBdd(left, right) != 1 || Ivy_FraigNodesAreEquivBdd(left, different) != 0 {
		t.Fatal("BDD checker disagrees")
	}
}
func TestFraigMuxAndXorCnfExhaustive(t *testing.T) {
	for phase := 0; phase < 8; phase++ {
		_, p := ivySatFixture(3)
		graph := p.pManFraig
		a, b, c := graph.vPis[0], graph.vPis[1], graph.vPis[2]
		root := Ivy_Mux(graph, Ivy_NotCond(a, int32(phase&1)), Ivy_NotCond(b, int32(phase&2)), Ivy_NotCond(c, int32(phase&4)))
		xor := Ivy_CanonExor(graph, a, b)
		Ivy_ObjCreatePo(graph, root)
		Ivy_ObjCreatePo(graph, xor)
		Ivy_FraigNodeAddToSolver(p, Ivy_Regular(root), Ivy_Regular(xor))
		for assignment := 0; assignment < 8; assignment++ {
			truth := ivyBalanceEvaluate(graph, assignment)
			var input []bsat.Literal
			for i, n := range graph.vPis {
				literal := ivyFraigLiteral(n)
				if assignment&(1<<i) == 0 {
					literal ^= 1
				}
				input = append(input, literal)
			}
			for i, edge := range []*Ivy_Obj_t{root, xor} {
				literal := ivyFraigLiteral(edge)
				if !truth[i] {
					literal ^= 1
				}
				correct := append(append([]bsat.Literal(nil), input...), literal)
				if p.pSat.Solve(correct...) != bsat.Satisfiable {
					t.Fatalf("correct assignment rejected phase%d assignment%d root%d", phase, assignment, i)
				}
				wrong := append(append([]bsat.Literal(nil), input...), literal^1)
				if p.pSat.Solve(wrong...) != bsat.Unsatisfiable {
					t.Fatal("CNF permits incorrect output")
				}
			}
		}
	}
}
func TestFraigPerformAssociativityReductionAndTruth(t *testing.T) {
	graph := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph)
	left, right := Ivy_And(graph, Ivy_And(graph, a, b), c), Ivy_And(graph, a, Ivy_And(graph, b, c))
	Ivy_ObjCreatePo(graph, left)
	Ivy_ObjCreatePo(graph, right)
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.nSimWords = 2
	result := Ivy_FraigPerform(graph, &params)
	if Ivy_ManNodeNum(result) >= Ivy_ManNodeNum(graph) {
		t.Fatal("equivalent nonstructural cones were not merged")
	}
	for assignment := 0; assignment < 8; assignment++ {
		old, new := ivyBalanceEvaluate(graph, assignment), ivyBalanceEvaluate(result, assignment)
		for i := range old {
			if old[i] != new[i] {
				t.Fatal("FRAIG changed truth")
			}
		}
	}
	if Ivy_ObjChild0(result.vPos[0]) != Ivy_ObjChild0(result.vPos[1]) {
		t.Fatal("equivalent outputs not shared")
	}
	if Ivy_ManCheck(graph) == 0 || Ivy_ManCheck(result) == 0 {
		t.Fatal("FRAIG damaged graph ownership")
	}
}
func TestFraigMiterSatisfiableAndUnsatisfiable(t *testing.T) {
	for _, satisfiable := range []bool{false, true} {
		graph := Ivy_ManStart()
		a, b, c := Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph)
		x, y := Ivy_And(graph, Ivy_And(graph, a, b), c), Ivy_And(graph, a, Ivy_And(graph, b, c))
		if satisfiable {
			y = Ivy_And(graph, a, c)
		}
		root := Ivy_Exor(graph, x, y)
		Ivy_ObjCreatePo(graph, root)
		var params Ivy_FraigParams_t
		Ivy_FraigParamsDefault(&params)
		params.FProve = 1
		params.NBTLimitMiter = 0
		result := Ivy_FraigMiter(graph, &params)
		status := Ivy_FraigMiterStatus(result)
		if satisfiable {
			if status != 0 {
				t.Fatal("SAT miter not detected")
			}
			model := result.pData.([]int32)
			assignment := 0
			for i, value := range model {
				if value != 0 {
					assignment |= 1 << i
				}
			}
			if !ivyBalanceEvaluate(graph, assignment)[0] {
				t.Fatal("miter model is invalid")
			}
		} else if status != 1 {
			t.Fatal("UNSAT miter not proved")
		}
	}
}
func TestFraigRandomReductionAndBddConsistency(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 19))
	for trial := 0; trial < 40; trial++ {
		p := Ivy_ManStart()
		var pool []*Ivy_Obj_t
		for i := 0; i < 4; i++ {
			pool = append(pool, Ivy_ObjCreatePi(p))
		}
		for i := 0; i < 25; i++ {
			a, b := pool[rng.IntN(len(pool))], pool[rng.IntN(len(pool))]
			pool = append(pool, Ivy_And(p, Ivy_NotCond(a, int32(rng.IntN(2))), Ivy_NotCond(b, int32(rng.IntN(2)))))
		}
		for i := 0; i < 4; i++ {
			Ivy_ObjCreatePo(p, pool[len(pool)-1-i])
		}
		Ivy_ManCleanup(p)
		var params Ivy_FraigParams_t
		Ivy_FraigParamsDefault(&params)
		params.nSimWords = 1
		result := Ivy_FraigPerform(p, &params)
		x, y := Ivy_ObjChild0(p.vPos[0]), Ivy_ObjChild0(p.vPos[1])
		equivalent := true
		for assignment := 0; assignment < 16; assignment++ {
			v := ivyBalanceEvaluate(p, assignment)
			if (v[0] != (Ivy_ObjFaninPhase(x) != 0)) != (v[1] != (Ivy_ObjFaninPhase(y) != 0)) {
				equivalent = false
			}
		}
		if (Ivy_FraigNodesAreEquivBdd(x, y) != 0) != equivalent {
			t.Fatal("random BDD checker differs from scalar normalized truth")
		}
		for assignment := 0; assignment < 16; assignment++ {
			a, b := ivyBalanceEvaluate(p, assignment), ivyBalanceEvaluate(result, assignment)
			for i := range a {
				if a[i] != b[i] {
					t.Fatalf("random FRAIG truth trial%d assignment%d", trial, assignment)
				}
			}
		}
	}
}

func TestFraigFalseSimulationCandidatesRefineBySatCounterexample(t *testing.T) {
	graph := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph), Ivy_ObjCreatePi(graph)
	root := Ivy_And(graph, Ivy_And(graph, a, b), c)
	Ivy_ObjCreatePo(graph, root)
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.nSimWords = 1
	p := Ivy_FraigStart(graph, &params)
	for _, input := range graph.vPis {
		Ivy_NodeAssignConst(p, input, 0)
	}
	Ivy_FraigSimulateOne(p)
	Ivy_FraigCreateClasses(p)
	Ivy_FraigSweep(p)
	if p.nSatCallsSat == 0 {
		t.Fatal("deliberately false simulation candidates were not checked by SAT")
	}
	for assignment := 0; assignment < 8; assignment++ {
		if ivyBalanceEvaluate(graph, assignment)[0] != ivyBalanceEvaluate(p.pManFraig, assignment)[0] {
			t.Fatal("SAT refinement changed truth")
		}
	}
	Ivy_FraigStop(p)
}
func TestFraigExtractedConeAndFreshSat(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	left, right := Ivy_And(p, Ivy_And(p, a, b), c), Ivy_And(p, a, Ivy_And(p, b, c))
	different := Ivy_And(p, a, c)
	var leaves []int32
	graph := Ivy_FraigExtractCone(p, left, right, &leaves)
	for assignment := 0; assignment < 8; assignment++ {
		values := make([]bool, len(leaves))
		for i, id := range leaves {
			values[i] = assignment&(1<<uint(id-1)) != 0
		}
		out, err := graph.Evaluate(values)
		if err != nil || out[0] {
			t.Fatal("extracted equal-cone miter is satisfiable")
		}
	}
	global := &Ivy_FraigMan_t{pPatWords: make([]uint32, 1)}
	if Ivy_FraigCheckCone(global, p, left, right, 0) != 1 || Ivy_FraigCheckCone(global, p, left, different, 0) != 0 {
		t.Fatal("fresh cone SAT status wrong")
	}
	for _, n := range p.vObjs {
		if n != nil && n.fMarkB != 0 {
			t.Fatal("cone extraction leaked marks")
		}
	}
}
func TestFraigProveSatOnlyDriver(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	left, right := Ivy_And(p, Ivy_And(p, a, b), c), Ivy_And(p, a, Ivy_And(p, b, c))
	Ivy_ObjCreatePo(p, Ivy_Exor(p, left, right))
	old := p
	status := Ivy_FraigProve(&p, &Prove_Params_t{})
	if status != 1 || p == old || !old.memoryStopped {
		t.Fatal("SAT-only prove did not replace/prove original manager")
	}
}

func TestFraigOriginalAbcAssociativityFixture(t *testing.T) {
	binary, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	fixture, err := filepath.Abs(filepath.Join("test_vectors", "fraig_associativity.blif"))
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "reference.aig")
	command := fmt.Sprintf("read_blif %s; strash; ifraig; write_aiger %s", fixture, output)
	if out, err := exec.Command(binary, "-c", command).CombinedOutput(); err != nil {
		t.Fatalf("original ABC fixture: %v\n%s", err, out)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := ioa.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	left, right := Ivy_And(p, Ivy_And(p, a, b), c), Ivy_And(p, a, Ivy_And(p, b, c))
	Ivy_ObjCreatePo(p, left)
	Ivy_ObjCreatePo(p, right)
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	result := Ivy_FraigPerform(p, &params)
	if int(Ivy_ManNodeNum(result)) != len(reference.ANDs) {
		t.Fatalf("Go nodes %d C nodes %d", Ivy_ManNodeNum(result), len(reference.ANDs))
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := make([]bool, reference.MaxVar+1)
		for i, literal := range reference.Inputs {
			values[literal>>1] = assignment&(1<<i) != 0
		}
		value := func(literal uint32) bool {
			v := values[literal>>1]
			if literal&1 != 0 {
				return !v
			}
			return v
		}
		for _, gate := range reference.ANDs {
			values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
		}
		got := ivyBalanceEvaluate(result, assignment)
		for i, literal := range reference.Outputs {
			if got[i] != value(literal) {
				t.Fatalf("source ABC truth differs at assignment %d", assignment)
			}
		}
	}
}

func TestFraigPreservesComplementedEquivalenceAndConstProof(t *testing.T) {
	source, p := ivySatFixture(3)
	defer Ivy_ManStop(source)
	g := p.pManFraig
	a, b, c := g.vPis[0], g.vPis[1], g.vPis[2]
	positive := Ivy_And(g, Ivy_And(g, a, b), c)
	negative := Ivy_Regular(Ivy_Or(g, Ivy_Not(a), Ivy_Not(Ivy_And(g, b, c))))
	// The regular De Morgan node may itself be the positive function; compare
	// expected normalized phase directly for every assignment.
	Ivy_ObjCreatePo(g, positive)
	Ivy_ObjCreatePo(g, negative)
	want := true
	for assignment := 0; assignment < 8; assignment++ {
		v := ivyBalanceEvaluate(g, assignment)
		if (v[0] != (positive.fPhase != 0)) != (v[1] != (negative.fPhase != 0)) {
			want = false
		}
	}
	if (Ivy_FraigNodesAreEquiv(p, positive, negative) == 1) != want {
		t.Fatal("SAT equivalence phase differs from scalar truth")
	}
	if (Ivy_FraigNodesAreEquivBdd(positive, negative) == 1) != want {
		t.Fatal("BDD equivalence phase differs from scalar truth")
	}
	contradiction := Ivy_And(g, Ivy_And(g, a, b), Ivy_Not(Ivy_And(g, a, Ivy_Or(g, b, c))))
	if Ivy_FraigNodeIsConst(p, Ivy_Regular(contradiction)) != 1 {
		t.Fatal("nonstructural contradiction not proved constant")
	}
}
func TestFraigConeRejectsExistingMarksWithoutChangingThem(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, b)
	a.fMarkB = 1
	var leaves []int32
	func() {
		defer func() {
			if recover() == nil {
				t.Error("marked cone accepted")
			}
		}()
		Ivy_FraigExtractCone(p, root, root, &leaves)
	}()
	if a.fMarkB != 1 || root.fMarkB != 0 {
		t.Fatal("rejected extraction modified existing marks")
	}
}
