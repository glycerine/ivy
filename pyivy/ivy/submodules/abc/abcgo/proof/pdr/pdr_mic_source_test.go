package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/hash"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// C returns immediately on an inconclusive MIC query, before restoring the
// removed literal to the caller's cube.
func TestSimpleMicTimeoutLeavesAttemptedLiteralRemovedAsInC(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(y)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	p.timeToStop = util.Abc_Clock() - 1
	cube := testPdrSet([]int32{0, 2}, nil)
	if got := ZPdr_ManSimpleMic(p, 1, &cube); got != -1 {
		t.Fatalf("expired query returned %d", got)
	}
	if cube.Lits[0] != -1 && cube.Lits[1] != -1 {
		t.Fatalf("source leaves the attempted literal removed: %v", cube.Lits)
	}
}

// Gate a pigeonhole contradiction on the selected frame-one next-state
// literals. Its first CTG check needs SAT search; the callback expires the
// deadline only after that query, making later frame queries inconclusive.
func pdrCtgLateTimeoutFixture(t *testing.T, literals []int32) (*Pdr_Man_t, *Pdr_Set_t, *Pdr_Set_t, *int32, *int) {
	t.Helper()
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(x)
	g.AddOutput(y)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.FCtgs = 1
	var p *Pdr_Man_t
	calls := 0
	pars.pFuncStop = func(int32) int32 {
		calls++
		p.timeToStop = util.Abc_Clock() - 1
		return 0
	}
	p = Pdr_ManStart(g, &pars, nil)
	t.Cleanup(func() { Pdr_ManStop(p) })
	for frame := int32(0); frame < 4; frame++ {
		Pdr_ManCreateSolver(p, frame)
	}
	cube := testPdrSet(literals, nil)
	pred := Pdr_SetDup(cube)
	gate := append([]int32(nil), vec.Vec_IntArray(Pdr_ManCubeToLits(p, 1, pred, 0, 1))...)
	solver := p.vSolvers[1]
	var vars [5][4]int32
	for pigeon := range vars {
		for hole := range vars[pigeon] {
			vars[pigeon][hole] = solver.NewVariable()
		}
	}
	add := func(lits ...int32) {
		clause := make([]bsat.Literal, 0, len(gate)+len(lits))
		for _, lit := range gate {
			clause = append(clause, lit^1)
		}
		clause = append(clause, lits...)
		if !solver.AddClause(clause...) {
			t.Fatal("gated pigeonhole fixture is inconsistent before assumptions")
		}
	}
	for pigeon := range vars {
		add(bsat.Positive(vars[pigeon][0]), bsat.Positive(vars[pigeon][1]), bsat.Positive(vars[pigeon][2]), bsat.Positive(vars[pigeon][3]))
	}
	for hole := 0; hole < 4; hole++ {
		for a := 0; a < 5; a++ {
			for b := a + 1; b < 5; b++ {
				add(bsat.Negative(vars[a][hole]), bsat.Negative(vars[b][hole]))
			}
		}
	}
	added := int32(1)
	return p, cube, pred, &added, &calls
}

func TestDownLateCtgTimeoutAssertsDuringMICAsInC(t *testing.T) {
	p, cube, pred, added, calls := pdrCtgLateTimeoutFixture(t, []int32{0, 2})
	result := int32(-99)
	defer func() {
		if got := recover(); got != "pdr: CTG MIC returned unknown" || *calls == 0 || p.nCubes != 0 || p.nCalls < 3 {
			t.Fatalf("expected C MIC assertion after a frame-push timeout: panic=%v result=%d queries=%d callbacks=%d cubes=%d", got, result, p.nCalls, *calls, p.nCubes)
		}
	}()
	result = ZPdr_ManDown(p, 2, &cube, pred, nil, nil, added)
}

func TestDownLateCtgTimeoutAssertsAfterAddingCubeAsInC(t *testing.T) {
	p, cube, pred, added, calls := pdrCtgLateTimeoutFixture(t, []int32{0})
	result := int32(-99)
	defer func() {
		if got := recover(); got != "pdr: post-CTG cube query returned unknown" || *calls == 0 || p.nCubes != 1 || p.nCalls < 3 {
			t.Fatalf("expected C post-CTG assertion after pushing cube: panic=%v result=%d queries=%d callbacks=%d cubes=%d", got, result, p.nCalls, *calls, p.nCubes)
		}
	}()
	result = ZPdr_ManDown(p, 2, &cube, pred, nil, nil, added)
}

// In C an inconclusive CTG check breaks out of the CTG loop; it then joins
// with the predecessor, and this keep-set makes the join fail with result 0.
func TestDownCtgTimeoutContinuesToJoinAsInC(t *testing.T) {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	for range inputs {
		g.NewInput()
	}
	g.AddOutput(g.Const0())
	for _, input := range inputs {
		g.AddOutput(input)
	}
	aig.Aig_ManSetRegNum(g, 3)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.FCtgs = 1
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	for i := int32(0); i < 3; i++ {
		Pdr_ManCreateSolver(p, i)
	}
	p.timeToStop = util.Abc_Clock() - 1
	cube := testPdrSet([]int32{0, 2}, nil)
	pred := testPdrSet([]int32{0, 4}, nil)
	keep := hash.Hash_IntAlloc(5)
	hash.Hash_IntWriteEntry(keep, 2, 1)
	added := int32(1)
	if got := ZPdr_ManDown(p, 2, &cube, pred, keep, nil, &added); got != 0 || cube != nil {
		t.Fatalf("got result=%d cube=%v, want failed join", got, cube)
	}
}
