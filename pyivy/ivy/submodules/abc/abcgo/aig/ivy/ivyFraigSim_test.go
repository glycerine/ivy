package ivy

import (
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFraigNormalizedSimulationAndFanouts(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	var roots []*Ivy_Obj_t
	for phase := 0; phase < 4; phase++ {
		roots = append(roots, Ivy_And(p, Ivy_NotCond(a, int32(phase&1)), Ivy_NotCond(b, int32(phase&2))))
	}
	roots = append(roots, Ivy_CanonExor(p, a, b))
	for _, n := range roots {
		Ivy_ObjCreatePo(p, n)
	}
	Ivy_ManStartFanout(p)
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.nSimWords = 1
	sim := ivyFraigStartSimulation(p, &params)
	defer ivyFraigStopSimulation(sim)
	Ivy_ObjSim(a).pData[0] = 0xaaaaaaaa
	Ivy_ObjSim(b).pData[0] = 0xcccccccc
	Ivy_FraigSimulateOne(sim)
	first := make([]uint32, len(roots))
	for i, n := range roots {
		word := Ivy_ObjSim(Ivy_Regular(n)).pData[0]
		first[i] = word
		for assignment := 0; assignment < 32; assignment++ {
			actual := ivyBalanceEvaluate(p, assignment&3)[i]
			normalized := actual != (Ivy_Regular(n).fPhase != 0)
			if (word&(1<<uint(assignment)) != 0) != normalized {
				t.Fatalf("root %d assignment %d", i, assignment)
			}
		}
	}
	Ivy_FraigSimulateOneSim(sim)
	for i, n := range roots {
		if Ivy_ObjSim(Ivy_Regular(n)).pData[0] != first[i] {
			t.Fatal("linked and graph simulation differ")
		}
	}
	if Ivy_ManCheckFanouts(p) == 0 {
		t.Fatal("FRAIG metadata overwrote real fanouts")
	}
}
func TestFraigClassesRefineAndPatternScores(t *testing.T) {
	p := Ivy_ManStart()
	inputs := make([]*Ivy_Obj_t, 6)
	for i := range inputs {
		inputs[i] = Ivy_ObjCreatePi(p)
	}
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.nSimWords = 1
	params.fPatScores = 1
	sim := ivyFraigStartSimulation(p, &params)
	defer ivyFraigStopSimulation(sim)
	for _, n := range inputs {
		Ivy_ObjSim(n).pData[0] = 3
	}
	Ivy_FraigCreateClasses(sim)
	if sim.lClasses.nItems != 1 || Ivy_FraigCountPairsClasses(sim) != 15 {
		t.Fatal("initial class partition incorrect")
	}
	for i, n := range inputs {
		Ivy_ObjSim(n).pData[0] = uint32(i/2 + 1)
	}
	changed := Ivy_FraigRefineClass_rec(sim, sim.lClasses.pHead)
	if changed != 2 || sim.lClasses.nItems != 3 || Ivy_FraigCountPairsClasses(sim) != 3 {
		t.Fatal("refined class partition incorrect")
	}
	previous := (*Ivy_Obj_t)(nil)
	for class := sim.lClasses.pHead; class != nil; class = Ivy_ObjEquivListNext(class) {
		if Ivy_ObjEquivListPrev(class) != previous || Ivy_FraigCountClassNodes(class) != 2 {
			t.Fatal("broken class links")
		}
		previous = class
	}
	if sim.lClasses.pTail != previous {
		t.Fatal("broken class tail")
	}
	sim.pPatScores[4] = 17
	Ivy_ObjSim(inputs[0]).pData[0] |= 1 << 4
	if Ivy_FraigSelectBestPat(sim) != 17 || sim.pPatWords[0]&1 == 0 {
		t.Fatal("best distinguishing pattern not selected")
	}
}
func TestFraigDistanceOneAndSimWords(t *testing.T) {
	p := Ivy_ManStart()
	for i := 0; i < 70; i++ {
		Ivy_ObjCreatePi(p)
	}
	params := Ivy_FraigParams_t{nSimWords: 2}
	sim := ivyFraigStartSimulation(p, &params)
	defer ivyFraigStopSimulation(sim)
	pattern := []uint32{0xaaaaaaaa, 0xcccccccc, 0x12}
	Ivy_FraigAssignDist1(sim, pattern)
	for i, n := range p.vPis {
		for bit := 0; bit < 64; bit++ {
			want := pattern[i/32]&(1<<uint(i%32)) != 0
			if i < 63 && bit == i+1 {
				want = !want
			}
			got := Ivy_ObjSim(n).pData[bit/32]&(1<<uint(bit%32)) != 0
			if got != want {
				t.Fatal("distance-one neighborhood wrong")
			}
		}
	}
}
func TestFraigSimulationOpcodesRandomReference(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 13))
	p := &Ivy_FraigMan_t{nSimWords: 4}
	for opcode := int32(0); opcode < 16; opcode++ {
		for round := 0; round < 100; round++ {
			a, b := &Ivy_FraigSim_t{pData: make([]uint32, 4)}, &Ivy_FraigSim_t{pData: make([]uint32, 4)}
			for i := range a.pData {
				a.pData[i] = rng.Uint32()
				b.pData[i] = rng.Uint32()
			}
			out := &Ivy_FraigSim_t{Type: opcode, pData: make([]uint32, 4), pFanin0: a, pFanin1: b}
			Ivy_NodeSimulateSim(p, out)
			for w := 0; w < 4; w++ {
				for bit := 0; bit < 32; bit++ {
					x, y := a.pData[w]&(1<<uint(bit)) != 0, b.pData[w]&(1<<uint(bit)) != 0
					if opcode&4 != 0 {
						x = !x
					}
					if opcode&2 != 0 {
						y = !y
					}
					want := x && y
					if opcode&8 != 0 {
						want = x != y
					}
					if opcode&1 != 0 {
						want = !want
					}
					if (out.pData[w]&(1<<uint(bit)) != 0) != want {
						t.Fatal("simulation opcode differs from scalar reference")
					}
				}
			}
		}
	}
}

func TestFraigOutputCounterexampleHonorsNormalizedPhase(t *testing.T) {
	for _, invert := range []bool{false, true} {
		p := Ivy_ManStart()
		a := Ivy_ObjCreatePi(p)
		edge := a
		if invert {
			edge = Ivy_Not(a)
		}
		Ivy_ObjCreatePo(p, edge)
		params := Ivy_FraigParams_t{nSimWords: 1}
		sim := ivyFraigStartSimulation(p, &params)
		Ivy_NodeAssignConst(sim, a, 0)
		Ivy_FraigSimulateOne(sim)
		got := Ivy_FraigCheckOutputSims(sim)
		if (got != 0) != invert {
			t.Fatal("output phase ignored in simulation check")
		}
		if invert {
			model := sim.pManFraig.pData.([]int32)
			if len(model) != 1 || model[0] != 0 {
				t.Fatal("reported counterexample does not satisfy complemented output")
			}
		}
		ivyFraigStopSimulation(sim)
	}
	p := Ivy_ManStart()
	Ivy_ObjCreatePo(p, p.pConst1)
	sim := ivyFraigStartSimulation(p, &Ivy_FraigParams_t{nSimWords: 1})
	defer ivyFraigStopSimulation(sim)
	if Ivy_FraigCheckOutputSims(sim) != 1 {
		t.Fatal("constant-one output missed by normalized simulation")
	}
}
func TestFraigClassSimulationSaturatesAndReleasesState(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	left := Ivy_And(p, Ivy_And(p, a, b), c)
	right := Ivy_And(p, a, Ivy_And(p, b, c))
	Ivy_ObjCreatePo(p, left)
	Ivy_ObjCreatePo(p, right)
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	params.nSimWords = 2
	sim := ivyFraigStartSimulation(p, &params)
	Ivy_FraigSimulate(sim)
	if Ivy_NodeCompareSims(sim, left, right) == 0 || Ivy_FraigCountPairsClasses(sim) < 1 {
		t.Fatal("equivalent reconvergent cones not in candidate classes")
	}
	ivyFraigStopSimulation(sim)
	for _, n := range p.vObjs {
		if n != nil && n.fraig != nil {
			t.Fatal("simulation state retained on graph")
		}
	}
	second := ivyFraigStartSimulation(p, &params)
	ivyFraigStopSimulation(second)
}
func TestFraigSimulationSourceCDifferential(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "src", "aig", "ivy", "ivyFraig.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	start := strings.Index(source, "\nvoid Ivy_NodeSimulateSim(")
	if start < 0 {
		t.Fatal("cannot locate source simulation function")
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		t.Fatal("cannot locate function end")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "original_sim.inc"), []byte(source[start:start+end+3]), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "reference")
	if out, err := exec.Command(compiler, "-std=c99", "-O2", "-I", dir, filepath.Join("test_vectors", "fraig_sim_reference.c"), "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile source reference: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewPCG(3, 101))
	var input strings.Builder
	var expected []uint32
	p := &Ivy_FraigMan_t{nSimWords: 1}
	for opcode := int32(0); opcode < 8; opcode++ {
		for trial := 0; trial < 1024; trial++ {
			a, b := rng.Uint32(), rng.Uint32()
			fmt.Fprintf(&input, "%d %08x %08x\n", opcode, a, b)
			sim := &Ivy_FraigSim_t{Type: opcode, pData: make([]uint32, 1), pFanin0: &Ivy_FraigSim_t{pData: []uint32{a}}, pFanin1: &Ivy_FraigSim_t{pData: []uint32{b}}}
			Ivy_NodeSimulateSim(p, sim)
			expected = append(expected, sim.pData[0])
		}
	}
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run reference: %v\n%s", err, out)
	}
	words := strings.Fields(string(out))
	if len(words) != len(expected) {
		t.Fatal("reference output count differs")
	}
	for i, want := range expected {
		got, err := strconv.ParseUint(words[i], 16, 32)
		if err != nil || uint32(got) != want {
			t.Fatalf("source simulation differs at %d: C=%s Go=%08x", i, words[i], want)
		}
	}
}

func TestFraigRepeatedStartRejectsWithoutChangingMappings(t *testing.T) {
	p := Ivy_ManStart()
	Ivy_ObjCreatePi(p)
	params := Ivy_FraigParams_t{nSimWords: 1}
	first := ivyFraigStartSimulation(p, &params)
	defer ivyFraigStopSimulation(first)
	mapped := p.pConst1.pEquiv
	state := p.pConst1.fraig
	func() {
		defer func() {
			if recover() == nil {
				t.Error("duplicate FRAIG ownership accepted")
			}
		}()
		ivyFraigStartSimulation(p, &params)
	}()
	if p.pConst1.pEquiv != mapped || p.pConst1.fraig != state {
		t.Fatal("failed startup mutated active state")
	}
}
