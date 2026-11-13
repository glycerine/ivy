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
