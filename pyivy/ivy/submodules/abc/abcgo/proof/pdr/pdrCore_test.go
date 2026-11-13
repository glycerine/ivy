package pdr

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestPdrDefaultParametersResetPriorState(t *testing.T) {
	params := Pdr_Par_t{nConfLimit: 91, fVerbose: 1, nFailOuts: 7}
	Pdr_ManSetDefaultParams(&params)
	if params.nRecycle != 300 || params.nFrameMax != 10000 || params.nRandomSeed != 91648253 || params.iFrame != -1 || params.fUseSupp != 1 || params.fUsePropOut != 1 || params.fSkipDown != 1 {
		t.Fatalf("unexpected defaults: %+v", params)
	}
	if params.nConfLimit != 0 || params.fVerbose != 0 || params.nFailOuts != 0 {
		t.Fatal("old parameter state survived reset")
	}
}

func TestPdrUnsupportedMonolithicUpdatedSimulation(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		g := aig.NewGraph()
		g.AddOutput(g.Const0())
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fMonoCnf, pars.fNewXSim = 1, 1
		var got int32
		if incremental {
			got = IPdr_ManSolve(g, &pars)
		} else {
			got = Pdr_ManSolve(g, &pars)
		}
		if got != -1 {
			t.Fatalf("incremental=%v: got %d, want undecided", incremental, got)
		}
	}
}

func TestPdrRejectsExistingMultiOutputWitnessVector(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		t.Run(fmt.Sprintf("incremental=%v", incremental), func(t *testing.T) {
			g := aig.NewGraph()
			g.AddOutput(g.Const0())
			g.SetSequentialCounterexamples(vec.Vec_PtrStart(1))
			var pars Pdr_Par_t
			Pdr_ManSetDefaultParams(&pars)
			pars.fSolveAll, pars.fSilent = 1, 1
			defer func() {
				if recover() == nil {
					t.Fatal("expected the C source's existing-vector assertion")
				}
			}()
			if incremental {
				IPdr_ManSolve(g, &pars)
			} else {
				Pdr_ManSolve(g, &pars)
			}
		})
	}
}

func TestPdrDisabledSimpleGeneralizationDoesNotProduceCube(t *testing.T) {
	var minimum *Pdr_Set_t
	if got := Pdr_ManGeneralize2(nil, 0, nil, &minimum); got != 0 || minimum != nil {
		t.Fatalf("disabled source function returned %d and %+v", got, minimum)
	}
}

func TestPdrSimpleGeneralizationLosesMultiLiteralCubeAsInC(t *testing.T) {
	g := aig.NewGraph()
	first, second := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(first)
	g.AddOutput(second)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSimpleGeneral = 1
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	cube := testPdrSet([]int32{0, 2}, nil)
	var minimum *Pdr_Set_t
	if got := Pdr_ManGeneralize(p, 1, cube, nil, &minimum); got != 1 || minimum != nil {
		t.Fatalf("got result %d and cube %+v; source caller discards the cube", got, minimum)
	}
}

func TestPdrContainmentScansRequestedFrames(t *testing.T) {
	makeSet := func(lits ...int32) *Pdr_Set_t {
		return Pdr_SetCreate(vec.Vec_IntAllocArrayCopy(lits, len(lits)), vec.Vec_IntAlloc(0))
	}
	p := &Pdr_Man_t{vSolvers: make([]*bsat.Solver, 3), vClauses: vec.Vec_VecStart(3)}
	vec.Vec_VecPush(p.vClauses, 1, makeSet(0, 2))
	vec.Vec_VecPush(p.vClauses, 2, makeSet(4))
	if Pdr_ManCheckContainment(p, 1, makeSet(0, 2, 6)) != 1 {
		t.Fatal("expected clause contained at frame 1")
	}
	if Pdr_ManCheckContainment(p, 2, makeSet(0, 2, 6)) != 0 {
		t.Fatal("searched a frame before requested start")
	}
	if Pdr_ManCheckContainment(p, 2, makeSet(4, 6)) != 1 {
		t.Fatal("expected clause contained at frame 2")
	}
}

func TestPdrLiteralPrioritySorts(t *testing.T) {
	prio := vec.Vec_IntAllocArrayCopy([]int32{7, 2, 9, 2}, 4)
	manager := &Pdr_Man_t{vPrio: prio}
	cube := &Pdr_Set_t{nLits: 4, Lits: []int32{0, 2, 4, 6}}
	if got := Pdr_ManSortByPriority(manager, cube); !reflect.DeepEqual(got, []int32{1, 3, 0, 2}) {
		t.Fatalf("priority order = %v", got)
	}
	literals := []int32{0, 2, 4, 6}
	Vec_IntSelectSortPrioReverseLit(literals, int32(len(literals)), prio)
	if !reflect.DeepEqual(literals, []int32{4, 0, 2, 6}) {
		t.Fatalf("reverse priority order = %v", literals)
	}
}

func TestPdrSolveSimpleSequentialCircuits(t *testing.T) {
	for _, tc := range []struct {
		name           string
		property, next string
		want           int32
	}{
		{"constant false", "false", "false", 1},
		{"constant true", "true", "false", 0},
		{"unreachable latch", "latch", "false", 1},
		{"latch after input", "latch", "input", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			in := g.NewInput()
			latch := g.NewInput()
			signals := map[string]aig.Edge{"input": in, "latch": latch, "true": g.Const1(), "false": g.Const0()}
			g.AddOutput(signals[tc.property])
			g.AddOutput(signals[tc.next])
			aig.Aig_ManSetRegNum(g, 1)
			var pars Pdr_Par_t
			Pdr_ManSetDefaultParams(&pars)
			pars.fSilent = 1
			got := Pdr_ManSolve(g, &pars)
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
			if got == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
				t.Fatal("missing or invalid counterexample")
			}
		})
	}
}

// Compare PDR with exhaustive reachability for every one-register transition
// and property function of a register and one primary input.
func TestPdrSolveAgainstExhaustiveReachability(t *testing.T) {
	for trans := 0; trans < 16; trans++ {
		for prop := 0; prop < 16; prop++ {
			g := aig.NewGraph()
			in := g.NewInput()
			latch := g.NewInput()
			truth := func(table int) aig.Edge {
				result := g.Const0()
				for row := 0; row < 4; row++ {
					if table&(1<<row) == 0 {
						continue
					}
					x, y := in, latch
					if row&1 == 0 {
						x = x.Not()
					}
					if row&2 == 0 {
						y = y.Not()
					}
					term := g.AND(x, y)
					result = g.AND(result.Not(), term.Not()).Not()
				}
				return result
			}
			g.AddOutput(truth(prop))
			g.AddOutput(truth(trans))
			aig.Aig_ManSetRegNum(g, 1)
			reached := [2]bool{true, false}
			bad := false
			for changed := true; changed; {
				changed = false
				for state := 0; state < 2; state++ {
					if !reached[state] {
						continue
					}
					for input := 0; input < 2; input++ {
						row := input + 2*state
						if prop&(1<<row) != 0 {
							bad = true
						}
						next := (trans >> row) & 1
						if !reached[next] {
							reached[next] = true
							changed = true
						}
					}
				}
			}
			var pars Pdr_Par_t
			Pdr_ManSetDefaultParams(&pars)
			pars.fSilent = 1
			got := Pdr_ManSolve(g, &pars)
			want := int32(1)
			if bad {
				want = 0
			}
			if got != want {
				t.Fatalf("transition %#x property %#x: got %d want %d", trans, prop, got, want)
			}
			if got == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
				t.Fatalf("transition %#x property %#x: invalid counterexample", trans, prop)
			}
		}
	}
}

func TestPdrSolveAllReportsEachOutput(t *testing.T) {
	for _, store := range []int32{0, 1} {
		g := aig.NewGraph()
		input := g.NewInput()
		latch := g.NewInput()
		g.AddOutput(g.Const0())
		g.AddOutput(g.Const1())
		g.AddOutput(latch)
		g.AddOutput(input)
		aig.Aig_ManSetRegNum(g, 1)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSolveAll = 1
		pars.fStoreCex = store
		pars.fSilent = 1
		pars.fNotVerbose = 1
		if got := Pdr_ManSolve(g, &pars); got != 0 {
			t.Fatalf("store %d: result %d", store, got)
		}
		if got := vec.Vec_IntArray(pars.vOutMap); !reflect.DeepEqual(got, []int32{1, 0, 0}) {
			t.Fatalf("store %d: outcomes %v", store, got)
		}
		models := g.SequentialCounterexamples()
		if models == nil || vec.Vec_PtrEntry(models, 0) != nil {
			t.Fatalf("store %d: bad models", store)
		}
		for i := 1; i <= 2; i++ {
			entry := vec.Vec_PtrEntry(models, i)
			if entry == nil {
				t.Fatalf("store %d: missing marker %d", store, i)
			}
			if store != 0 {
				cex, ok := entry.(*util.Abc_Cex_t)
				if !ok || saig.Saig_ManVerifyCex(g, cex) != 1 {
					t.Fatalf("invalid counterexample %d: %v", i, entry)
				}
			}
		}
	}
}

func TestPdrSolveTwoRegisterReachability(t *testing.T) {
	seed := uint32(0x31415926)
	nextRandom := func() int { seed ^= seed << 13; seed ^= seed >> 17; seed ^= seed << 5; return int(seed & 255) }
	for trial := 0; trial < 100; trial++ {
		tables := [3]int{nextRandom(), nextRandom(), nextRandom()}
		g := aig.NewGraph()
		vars := [3]aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		truth := func(table int) aig.Edge {
			r := g.Const0()
			for row := 0; row < 8; row++ {
				if table&(1<<row) == 0 {
					continue
				}
				term := g.Const1()
				for i, v := range vars {
					if row&(1<<i) == 0 {
						v = v.Not()
					}
					term = g.AND(term, v)
				}
				r = g.AND(r.Not(), term.Not()).Not()
			}
			return r
		}
		g.AddOutput(truth(tables[2]))
		g.AddOutput(truth(tables[0]))
		g.AddOutput(truth(tables[1]))
		aig.Aig_ManSetRegNum(g, 2)
		reached := [4]bool{true}
		bad := false
		for changed := true; changed; {
			changed = false
			for state := 0; state < 4; state++ {
				if !reached[state] {
					continue
				}
				for input := 0; input < 2; input++ {
					row := input + 2*state
					if tables[2]&(1<<row) != 0 {
						bad = true
					}
					next := ((tables[0] >> row) & 1) | (((tables[1] >> row) & 1) << 1)
					if !reached[next] {
						reached[next] = true
						changed = true
					}
				}
			}
		}
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent = 1
		if trial%2 == 0 {
			pars.fSkipDown = 0
			pars.fCtgs = 1
		}
		if trial%3 == 0 {
			pars.fTwoRounds = 1
		}
		if trial%5 == 0 {
			pars.fFlopOrder = 1
		}
		if trial%7 == 0 {
			pars.fUseAbs = 1
			pars.fUseSimpleRef = 1
		}
		got := Pdr_ManSolve(g, &pars)
		want := int32(1)
		if bad {
			want = 0
		}
		if got != want {
			t.Fatalf("trial %d tables %v: got %d want %d", trial, tables, got, want)
		}
		if got == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
			t.Fatalf("trial %d tables %v: invalid counterexample", trial, tables)
		}
	}
}

func TestPdrSolveAdditionalModesAgainstReachability(t *testing.T) {
	testAdditionalPdrModes(t, false, 512)
}

func TestIncrementalPdrAdditionalModesAgainstReachability(t *testing.T) {
	testAdditionalPdrModes(t, true, 128)
}

func TestPdrShortestTraceForLatchChain(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		g := aig.NewGraph()
		first := g.NewInput()
		second := g.NewInput()
		g.AddOutput(second)
		g.AddOutput(g.Const1())
		g.AddOutput(first)
		aig.Aig_ManSetRegNum(g, 2)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent, pars.fShortest = 1, 1
		var status int32
		if incremental {
			pars.nFrameMax = 1
			status = IPdr_ManSolve(g, &pars)
		} else {
			status = Pdr_ManSolve(g, &pars)
		}
		cex := g.SequentialCounterexample()
		if status != 0 || cex == nil || cex.IFrame != 2 || saig.Saig_ManVerifyCex(g, cex) != 1 {
			t.Fatalf("incremental=%v: status %d counterexample %+v", incremental, status, cex)
		}
	}
}

func testAdditionalPdrModes(t *testing.T, incremental bool, trials int) {
	t.Helper()
	seed := uint32(0x31415926)
	nextRandom := func() int { seed ^= seed << 13; seed ^= seed >> 17; seed ^= seed << 5; return int(seed & 255) }
	for trial := 0; trial < trials; trial++ {
		tables := [3]int{nextRandom(), nextRandom(), nextRandom()}
		g := aig.NewGraph()
		vars := [3]aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		truth := func(table int) aig.Edge {
			r := g.Const0()
			for row := 0; row < 8; row++ {
				if table&(1<<row) == 0 {
					continue
				}
				term := g.Const1()
				for i, v := range vars {
					if row&(1<<i) == 0 {
						v = v.Not()
					}
					term = g.AND(term, v)
				}
				r = g.AND(r.Not(), term.Not()).Not()
			}
			return r
		}
		g.AddOutput(truth(tables[2]))
		g.AddOutput(truth(tables[0]))
		g.AddOutput(truth(tables[1]))
		aig.Aig_ManSetRegNum(g, 2)
		reached := [4]bool{true}
		bad := false
		for changed := true; changed; {
			changed = false
			for state := 0; state < 4; state++ {
				if !reached[state] {
					continue
				}
				for input := 0; input < 2; input++ {
					row := input + 2*state
					if tables[2]&(1<<row) != 0 {
						bad = true
					}
					next := ((tables[0] >> row) & 1) | (((tables[1] >> row) & 1) << 1)
					if !reached[next] {
						reached[next] = true
						changed = true
					}
				}
			}
		}
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent = 1
		pars.fMonoCnf = int32(trial % 2)
		pars.fNewXSim = int32((trial / 2) % 2)
		pars.fUseAbs = int32((trial / 4) % 2)
		if pars.fNewXSim != 0 && (pars.fMonoCnf != 0 || pars.fUseAbs != 0) {
			continue // ABC's updated-simulation combinations can assert or crash.
		}
		pars.fUseSimpleRef = 0
		pars.fShortest = int32((trial / 8) % 2)
		pars.fFlopPrio = int32((trial / 16) % 2)
		pars.fShiftStart = int32((trial / 32) % 2)
		pars.fReuseProofOblig = int32((trial / 64) % 2)
		pars.fSkipDown = int32((trial / 128) % 2)
		pars.fCtgs = int32((trial / 256) % 2)
		var got int32
		if incremental {
			pars.nFrameMax = 1
			got = IPdr_ManSolve(g, &pars)
		} else {
			got = Pdr_ManSolve(g, &pars)
		}
		want := int32(1)
		if bad {
			want = 0
		}
		if got != want {
			t.Fatalf("trial %d tables %v: got %d want %d", trial, tables, got, want)
		}
		if got == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
			t.Fatalf("trial %d tables %v: invalid counterexample", trial, tables)
		}
	}
}

func TestPdrHonorsStopCallback(t *testing.T) {
	g := aig.NewGraph()
	in := g.NewInput()
	latch := g.NewInput()
	g.AddOutput(latch)
	g.AddOutput(in)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent = 1
	calls := 0
	pars.RunId = 42
	pars.pFuncStop = func(id int32) int32 {
		if id != 42 {
			t.Fatalf("callback id %d", id)
		}
		calls++
		return 1
	}
	if got := Pdr_ManSolve(g, &pars); got != -1 {
		t.Fatalf("got %d, want unknown", got)
	}
	if calls == 0 {
		t.Fatal("stop callback was not invoked")
	}
}
