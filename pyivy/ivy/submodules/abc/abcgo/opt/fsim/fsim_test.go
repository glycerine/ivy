package fsim

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"reflect"
	"testing"
)

func TestFrontierWordsAgainstScalar(t *testing.T) {
	r := rand.New(rand.NewSource(427))
	for trial := 0; trial < 60; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.Const0(), g.Const1(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 15; i++ {
			a, b := pool[r.Intn(len(pool))], pool[r.Intn(len(pool))]
			if r.Intn(2) != 0 {
				a = a.Not()
			}
			if i%3 == 0 {
				pool = append(pool, g.NewXor(a, b))
			} else {
				pool = append(pool, g.AND(a, b))
			}
		}
		for i := 0; i < 3; i++ {
			g.AddOutput(pool[len(pool)-1-i])
		}
		p, err := Fsim_ManCreate(g)
		if err != nil {
			t.Fatal(err)
		}
		cis := make([][]uint32, 4)
		for i := range cis {
			cis[i] = []uint32{r.Uint32(), r.Uint32()}
		}
		var reference [][]uint32
		for _, compressed := range []bool{false, true} {
			if err = Fsim_ManFront(p, compressed); err != nil {
				t.Fatal(err)
			}
			if err = Fsim_ManVerifyFront(p); err != nil {
				t.Fatal(err)
			}
			out, err := p.Round(cis)
			if err != nil {
				t.Fatal(err)
			}
			if reference != nil && !reflect.DeepEqual(reference, out) {
				t.Fatal("compressed output differs")
			}
			reference = out
			for pattern := 0; pattern < 64; pattern++ {
				input := make([]bool, 4)
				for i := range input {
					input[i] = cis[i][pattern>>5]&(1<<uint(pattern&31)) != 0
				}
				want, err := g.Evaluate(input)
				if err != nil {
					t.Fatal(err)
				}
				for output, v := range want {
					got := out[output][pattern>>5]&(1<<uint(pattern&31)) != 0
					if got != v {
						t.Fatal("frontier result differs", trial, pattern, output)
					}
				}
			}
		}
	}
}
func TestSequentialWitnessReplayAndOwnership(t *testing.T) {
	g := aig.NewGraph()
	a, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	for _, compressed := range []bool{false, true} {
		result, err := RunSimulation(g, &Fsim_ParSim_t{Words: 2, Iterations: 5, CheckMiter: true, CompressAIG: compressed})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Found || result.Frames != 2 || result.Counterexample.IFrame != 1 {
			t.Fatal("missing delayed witness", result)
		}
		if err = verifyCounter(g, result.Counterexample); err != nil {
			t.Fatal(err)
		}
		owned := g.SequentialCounterexample()
		result.Counterexample.PData[0] ^= ^uint32(0)
		if reflect.DeepEqual(owned.PData, result.Counterexample.PData) {
			t.Fatal("borrowed graph witness aliases result")
		}
	}
}
func TestMiterPatternIndicesAndTernaryCycles(t *testing.T) {
	if Fsim_ManSimInfoIsZero([]uint32{1}) != 0 || Fsim_ManSimInfoIsZero([]uint32{0, 0x80000000}) != 63 {
		t.Fatal("miter word index is shifted")
	}
	g := aig.NewGraph()
	q := g.NewInput()
	g.AddOutput(q)
	g.AddOutput(q.Not())
	aig.Aig_ManSetRegNum(g, 1)
	result, err := RunTernary(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.States) != 3 || result.CycleStart != 0 || Fsim_ManTerSimInfoGet(result.States[0], 0) != FSIM_ZER || Fsim_ManTerSimInfoGet(result.States[1], 0) != FSIM_ONE {
		t.Fatal("ternary oscillator", result)
	}
	if _, err = RunTernary(g, 2); err == nil {
		t.Fatal("state limit ignored")
	}
	if Aig_XsimAndCond(FSIM_ZER, 0, FSIM_UND, 0) != FSIM_ZER {
		t.Fatal("controlling zero lost to unknown")
	}
}
func TestTernaryCIOrderAndControllingZero(t *testing.T) {
	g := aig.NewGraph()
	a, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(g.AND(q, a))
	aig.Aig_ManSetRegNum(g, 1)
	p, err := Fsim_ManCreate(g)
	if err != nil {
		t.Fatal(err)
	}
	if p.CIOrder[0] != 1 {
		t.Fatal("fixture did not shuffle inputs")
	}
	result, err := RunTernary(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.States) != 2 || Fsim_ManTerSimInfoGet(result.States[1], 0) != FSIM_ZER {
		t.Fatal("original RO order or controlling-zero abstraction lost")
	}
}
func TestSmallFrontierAndConstantSwitching(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.AddOutput(g.Const1())
	if err := Fsim_ManTest(g); err != nil {
		t.Fatal(err)
	}
	constant := aig.NewGraph()
	constant.AddOutput(constant.Const1())
	result, err := Fsim_ManSwitchSimulate(constant, &Fsim_ParSwitch_t{Words: 2, Iterations: 3, ProbabilityOne: true, ProbabilityTransition: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.One[0] != 1 || result.Transition[0] != 0 {
		t.Fatal("wrong constant activity", result)
	}
}
func TestSwitchingToggleRegister(t *testing.T) {
	g := aig.NewGraph()
	q := g.NewInput()
	g.AddOutput(q)
	g.AddOutput(q.Not())
	aig.Aig_ManSetRegNum(g, 1)
	result, err := Fsim_ManSwitchSimulate(g, &Fsim_ParSwitch_t{Words: 2, Iterations: 10, Prefix: 2, ProbabilityOne: true, ProbabilityTransition: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.One[q.NodeID()] != 0.5 || result.Transition[q.NodeID()] != 1 || result.OneFrames != 8 || result.TransitionFrames != 8 {
		t.Fatal("toggle activity", result)
	}
}
func TestCodecRejectsTruncatedOrZeroNodes(t *testing.T) {
	for _, data := range [][]byte{{0x80}, {byte(aig.And), 0, 0}, {byte(aig.Input)}} {
		offset, previous := 0, 0
		if _, err := decodeOperation(data, &offset, &previous); err == nil {
			t.Errorf("invalid compressed operation accepted: %v", data)
		}
	}
}
func TestBufferAndOutputAliases(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{a.Not(), aig.InvalidEdge}})
	co := g.AddOutput(buffer.Not())
	gate := g.NewXor(co, b)
	g.AddOutput(gate)
	p, err := Fsim_ManCreate(g)
	if err != nil {
		t.Fatal(err)
	}
	for _, compressed := range []bool{false, true} {
		if err = Fsim_ManFront(p, compressed); err != nil {
			t.Fatal(err)
		}
		out, err := p.Round([][]uint32{{0xaaaaaaaa}, {0xcccccccc}})
		if err != nil {
			t.Fatal(err)
		}
		if out[0][0] != 0xaaaaaaaa || out[1][0] != (0xaaaaaaaa^0xcccccccc) {
			t.Fatal("buffer/output alias phase lost", out)
		}
	}
}
func TestOrdinarySimulationPreservesExistingWitness(t *testing.T) {
	g := aig.NewGraph()
	a, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	if _, err := RunSimulation(g, &Fsim_ParSim_t{Words: 1, Iterations: 3, CheckMiter: true}); err != nil {
		t.Fatal(err)
	}
	model := g.SequentialCounterexample()
	if model == nil {
		t.Fatal("fixture has no witness")
	}
	if _, err := RunSimulation(g, &Fsim_ParSim_t{Words: 1, Iterations: 2}); err != nil {
		t.Fatal(err)
	}
	if g.SequentialCounterexample() != model {
		t.Error("ordinary simulation cleared an unrelated witness")
	}
}
func TestConstantFirstCompressedProgram(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AND(a, b)
	g.AddOutput(g.Const1())
	p, err := Fsim_ManCreate(g)
	if err != nil {
		t.Fatal(err)
	}
	if p.ops[2].Kind != aig.Output {
		t.Fatal("fixture did not start with constant CO")
	}
	if err = Fsim_ManFront(p, true); err != nil {
		t.Fatal(err)
	}
	if err = Fsim_ManVerifyFront(p); err != nil {
		t.Fatal(err)
	}
	out, err := p.Round([][]uint32{{1}, {2}})
	if err != nil || out[0][0] != ^uint32(0) {
		t.Fatal("constant-first compressed program", out, err)
	}
}
func TestSwitchingProbabilityFlagsAndRate(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	result, err := Fsim_ManSwitchSimulate(g, &Fsim_ParSwitch_t{Words: 4, Iterations: 128, Prefix: 1, RandomPI: 1, ProbabilityTransition: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.One != nil || result.Transition[a.NodeID()] < 0.23 || result.Transition[a.NodeID()] > 0.27 {
		t.Fatal("wrong quarter-rate switching", result.Transition)
	}
	result, err = Fsim_ManSwitchSimulate(g, &Fsim_ParSwitch_t{Words: 1, Iterations: 4, ProbabilityOne: true})
	if err != nil || result.Transition != nil {
		t.Fatal("probability flags ignored", err)
	}
	state := make([]uint32, 1)
	Fsim_ManTerSimInfoSet(state, 2, FSIM_UND)
	Fsim_ManTerSimInfoSet(state, 2, FSIM_ZER)
	if Fsim_ManTerSimInfoGet(state, 2) != FSIM_ZER {
		t.Fatal("ternary overwrite retained stale phase bit")
	}
}
