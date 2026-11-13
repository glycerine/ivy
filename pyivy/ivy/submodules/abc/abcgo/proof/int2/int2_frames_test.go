package int2

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCreateFramesPreservesComplementedAndFanins(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a^1, b)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cos, 0)
	Int2_ManCreateFrames(m, 0, cos)
	av := Int2_ManCheckFrames(m, 0, a>>1)
	bv := Int2_ManCheckFrames(m, 0, b>>1)
	output := Int2_ManCheckFrames(m, 0, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
	for x := int32(0); x < 2; x++ {
		for y := int32(0); y < 2; y++ {
			for z := int32(0); z < 2; z++ {
				assumptions := []int32{av ^ (1 - x), bv ^ (1 - y), output ^ (1 - z)}
				status := sat_solver_solve(m.pSatPref, assumptions, 0, 0, 0, 0)
				if (status == l_True) != (z == ((1 - x) & y)) {
					t.Fatalf("x=%d y=%d z=%d status=%d", x, y, z, status)
				}
			}
		}
	}
}

func TestCreateFramesMapsConstantAtEveryFrame(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cos, 0)
	for frame := int32(0); frame < 2; frame++ {
		Int2_ManCreateFrames(m, frame, cos)
		if got := Int2_ManCheckFrames(m, frame, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0))); got != 0 {
			t.Fatalf("frame %d constant output = %d", frame, got)
		}
	}
}

func TestCreateFramesAllocatesDirectInputSATVariable(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cos, 0)
	Int2_ManCreateFrames(m, 0, cos)
	if got := Int2_ManCheckBmc(m, nil); got != 0 {
		t.Fatalf("input-driven bad output: status %d, want counterexample", got)
	}
}

func TestCheckInitConstantTransitions(t *testing.T) {
	for _, tc := range []struct{ next, want int32 }{{0, 1}, {1, 0}} {
		g := gia.Gia_ManStart(8)
		gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, 0)
		gia.Gia_ManAppendCo(g, tc.next)
		gia.Gia_ManSetRegNum(g, 1)
		if got := Int2_ManCheckInit(g); got != tc.want {
			t.Fatalf("constant next-state %d: got %d want %d", tc.next, got, tc.want)
		}
	}
}

func TestFrameUnrollingTracksLatestInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g) // initial register state
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManSetRegNum(g, 1)
	for _, build := range []struct {
		name string
		fn   func(*Gia_Man_t, int32) *Gia_Man_t
	}{
		{"frame", func(g *Gia_Man_t, n int32) *Gia_Man_t { return Int2_ManFrameInit(g, n, 0) }},
		{"unroll", Int2_ManUnroll},
	} {
		t.Run(build.name, func(t *testing.T) {
			out := build.fn(g, 2)
			defer gia.Gia_ManStop(out)
			if got := gia.Gia_ManPiNum(out); got != 2 {
				t.Fatalf("inputs = %d, want 2", got)
			}
			for mask := 0; mask < 4; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0}
				got := int2EvalGraph(out, inputs, nil, make([]bool, gia.Gia_ManObjNum(out)), nil)
				if got != inputs[1] {
					t.Fatalf("inputs %v => %v, want latest input", inputs, got)
				}
			}
		})
	}
}

func TestSetupBmcSolverConstrainsFinalState(t *testing.T) {
	for _, next := range []int32{0, 1} {
		g := gia.Gia_ManStart(8)
		gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, 0)
		gia.Gia_ManAppendCo(g, next)
		gia.Gia_ManSetRegNum(g, 1)
		solver := Int2_ManSetupBmcSolver(g, 1)
		if solver == nil {
			t.Fatalf("constant %d: no solver", next)
		}
		out := sat_solver_nvars(solver) - 1
		for value := int32(0); value < 2; value++ {
			got := sat_solver_solve(solver, []int32{sat_literal(out, 1-value)}, 0, 0, 0, 0)
			if (got == l_True) != (value == next) {
				t.Fatalf("next=%d value=%d status=%d", next, value, got)
			}
		}
	}
}

func TestPreparePrefixAndSuffixUseMappedCNF(t *testing.T) {
	g := testInt2InputToBadMachine()
	var prefixMap *Vec_Int_t
	prefix := Int2_ManPreparePrefix(g, 1, &prefixMap)
	if prefix == nil || prefixMap == nil || vec.Vec_IntSize(prefixMap) != 1 {
		t.Fatal("prefix solver/map missing")
	}
	variable := vec.Vec_IntEntry(prefixMap, 0)
	for value := int32(0); value < 2; value++ {
		if sat_solver_solve(prefix, []int32{sat_literal(variable, 1-value)}, 0, 0, 0, 0) != l_True {
			t.Fatalf("prefix should allow next-state value %d", value)
		}
	}
	var suffixMap *Vec_Int_t
	var suffix *Gia_Man_t
	images := vec.Vec_IntAlloc(0)
	suffixSAT := Int2_ManPrepareSuffix(g, nil, images, &suffixMap, &suffix)
	if suffixSAT == nil || suffix == nil || suffixMap == nil || vec.Vec_IntSize(suffixMap) != 2 {
		t.Fatal("suffix solver/graph/map missing")
	}
	defer gia.Gia_ManStop(suffix)
	stateVariable := vec.Vec_IntEntry(suffixMap, 1)
	if sat_solver_solve(suffixSAT, []int32{sat_literal(stateVariable, 0)}, 0, 0, 0, 0) != l_True {
		t.Fatal("bad state should satisfy suffix")
	}
	if sat_solver_solve(suffixSAT, []int32{sat_literal(stateVariable, 1)}, 0, 0, 0, 0) != l_False {
		t.Fatal("safe state should not satisfy suffix")
	}
}

func TestPrepareSuffixBlocksStateCubeUsingMappedVariable(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g) // primary input precedes the register output
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1) // property is true for either state
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	images := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 0}, 3) // block state=1
	var coMap *Vec_Int_t
	var suffix *Gia_Man_t
	solver := Int2_ManPrepareSuffix(g, nil, images, &coMap, &suffix)
	if solver == nil || suffix == nil || coMap == nil {
		t.Fatal("suffix solver/graph/map missing")
	}
	defer gia.Gia_ManStop(suffix)
	variable := vec.Vec_IntEntry(coMap, 1)
	for value := int32(0); value < 2; value++ {
		status := sat_solver_solve(solver, []int32{sat_literal(variable, 1-value)}, 0, 0, 0, 0)
		if (status == l_True) != (value == 0) {
			t.Fatalf("state=%d: status %d", value, status)
		}
	}
}

func TestPrepareSuffixBlocksMultipleCubes(t *testing.T) {
	g := gia.Gia_ManStart(16)
	gia.Gia_ManAppendCi(g) // unrelated primary input
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManSetRegNum(g, 2)
	// Block x=1,y=0 and x=0,y=1. Their complement is x==y.
	images := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 0, 3, 2, 1, 2}, 7)
	var coMap *Vec_Int_t
	var suffix *Gia_Man_t
	solver := Int2_ManPrepareSuffix(g, nil, images, &coMap, &suffix)
	if solver == nil || coMap == nil || suffix == nil {
		t.Fatal("suffix solver/graph/map missing")
	}
	defer gia.Gia_ManStop(suffix)
	for xv := int32(0); xv < 2; xv++ {
		for yv := int32(0); yv < 2; yv++ {
			assumptions := []int32{
				sat_literal(vec.Vec_IntEntry(coMap, 1), 1-xv),
				sat_literal(vec.Vec_IntEntry(coMap, 2), 1-yv),
			}
			status := sat_solver_solve(solver, assumptions, 0, 0, 0, 0)
			if (status == l_True) != (xv == yv) {
				t.Fatalf("x=%d y=%d: status %d", xv, yv, status)
			}
		}
	}
}

func TestCheckInitAgainstExhaustiveTransitions(t *testing.T) {
	// Every pair of Boolean next-state functions over two registers.
	for table0 := int32(0); table0 < 16; table0++ {
		for table1 := int32(0); table1 < 16; table1++ {
			g := gia.Gia_ManStart(32)
			a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			gia.Gia_ManHashAlloc(g)
			build := func(table int32) int32 {
				result := int32(0)
				for mask := int32(0); mask < 4; mask++ {
					if table&(1<<mask) == 0 {
						continue
					}
					left, right := a^((mask&1)^1), b^(((mask>>1)&1)^1)
					result = gia.Gia_ManHashOr(g, result, gia.Gia_ManHashAnd(g, left, right))
				}
				return result
			}
			n0, n1 := build(table0), build(table1)
			gia.Gia_ManAppendCo(g, 0)
			gia.Gia_ManAppendCo(g, n0)
			gia.Gia_ManAppendCo(g, n1)
			gia.Gia_ManHashStop(g)
			gia.Gia_ManSetRegNum(g, 2)
			want := false
			for mask := int32(0); mask < 4; mask++ {
				if (table0|table1)&(1<<mask) == 0 {
					want = true
				}
			}
			if got := Int2_ManCheckInit(g) != 0; got != want {
				t.Fatalf("tables %04b %04b: got %v want %v", table0, table1, got, want)
			}
			gia.Gia_ManStop(g)
		}
	}
}

func TestFrameInitMatchesSequentialSimulation(t *testing.T) {
	g := gia.Gia_ManStart(32)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	next0 := gia.Gia_ManAppendMuxReal(g, a, x^1, b)
	next1 := gia.Gia_ManAppendXorReal(g, y, next0)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendBuf(g, x))
	gia.Gia_ManAppendCo(g, next0)
	gia.Gia_ManAppendCo(g, next1)
	gia.Gia_ManSetRegNum(g, 2)
	frames := Int2_ManFrameInit(g, 3, 0)
	defer gia.Gia_ManStop(frames)
	for mask := 0; mask < 64; mask++ {
		inputs := make([]bool, 6)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		state := []bool{false, false}
		for frame := 0; frame < 3; frame++ {
			out, err := g.Evaluate([]bool{inputs[2*frame], inputs[2*frame+1], state[0], state[1]})
			if err != nil {
				t.Fatal(err)
			}
			state[0], state[1] = out[1], out[2]
		}
		out, err := frames.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != state[0] || out[1] != state[1] {
			t.Fatalf("input=%v got=%v want=%v", inputs, out, state)
		}
	}
}

func TestCreateFramesCanStartAtLaterFrame(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	Int2_ManCreateFrames(m, 2, cos)
	if got := sat_solver_solve(m.pSatPref, []int32{0}, 0, 0, 0, 0); got != l_False {
		t.Fatalf("constant-zero SAT variable may be true: status %d", got)
	}
}

func TestCheckBmcHandlesUnconstructedOutputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	Int2_ManCreateFrames(m, 0, cos)
	if got := Int2_ManCheckBmc(m, nil); got != -1 {
		t.Fatalf("unconstructed output result %d, want undecided", got)
	}
}

func TestCheckBmcFindsFailureAfterUnconstructedOutput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	Int2_ManCreateFrames(m, 0, cos)
	if got := Int2_ManCheckBmc(m, nil); got != 0 {
		t.Fatalf("known failing output result %d, want failure", got)
	}
}

func TestZeroFrameBuildersReturnInitialState(t *testing.T) {
	g := gia.Gia_ManStart(12)
	pi, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, pi^1)
	gia.Gia_ManSetRegNum(g, 1)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, pi>>1), 1)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, state>>1), 1)
	gia.Gia_ObjSetValue(gia.Gia_ManRi(g, 0), 1)
	for _, builder := range []struct {
		name string
		run  func(*Gia_Man_t) *Gia_Man_t
	}{
		{"unroll", func(p *Gia_Man_t) *Gia_Man_t { return Int2_ManUnroll(p, 0) }},
		{"frame init", func(p *Gia_Man_t) *Gia_Man_t { return Int2_ManFrameInit(p, 0, 0) }},
	} {
		out := builder.run(g)
		result, err := out.Evaluate(nil)
		if err != nil {
			t.Fatalf("%s: %v", builder.name, err)
		}
		if len(result) != 1 || result[0] {
			t.Fatalf("%s: got %v, want zero initial state", builder.name, result)
		}
		gia.Gia_ManStop(out)
	}
}

func TestDupInitMatchesMuxedStateTransitions(t *testing.T) {
	g := gia.Gia_ManStart(24)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	gate := gia.Gia_ManAppendMuxReal(g, a, b^1, x)
	next := gia.Gia_ManAppendXorReal(g, y, gate)
	gia.Gia_ManAppendCo(g, gate^1)
	gia.Gia_ManAppendCo(g, next)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendBuf(g, b^1))
	gia.Gia_ManSetRegNum(g, 2)
	dup := Int2_ManDupInit(g, 0)
	defer gia.Gia_ManStop(dup)
	for mask := 0; mask < 32; mask++ {
		a0, b0, ctrl, x0, y0 := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0
		old, err := g.Evaluate([]bool{a0, b0, x0, y0})
		if err != nil {
			t.Fatal(err)
		}
		got, err := dup.Evaluate([]bool{a0, b0, ctrl, x0, y0})
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{old[0], old[1], old[2]}
		if ctrl {
			want[1], want[2] = x0, y0
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mask %d got %v want %v", mask, got, want)
			}
		}
	}
}

func TestPreparePrefixAtFrameZeroConstrainsInitialState(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManSetRegNum(g, 1)
	var mapping *Vec_Int_t
	s := Int2_ManPreparePrefix(g, 0, &mapping)
	if s == nil || mapping == nil || vec.Vec_IntSize(mapping) != 1 {
		t.Fatal("missing prefix or state map")
	}
	v := vec.Vec_IntEntry(mapping, 0)
	for value := int32(0); value < 2; value++ {
		status := sat_solver_solve(s, []int32{sat_literal(v, 1-value)}, 0, 0, 0, 0)
		if (status == l_True) != (value == 0) {
			t.Fatalf("state %d returned %d", value, status)
		}
	}
}
