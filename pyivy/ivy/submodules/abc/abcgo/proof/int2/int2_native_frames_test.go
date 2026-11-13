package int2

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCreateFramesNativeGates(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for polarity := int32(0); polarity < 8; polarity++ {
			g := gia.Gia_ManStart(32)
			a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCi(g) // register
			var output int32
			switch kind {
			case "xor":
				output = gia.Gia_ManAppendXorReal(g, a^(polarity&1), b^((polarity>>1)&1))
			case "mux":
				g.EnableMuxes()
				output = gia.Gia_ManAppendMuxReal(g, a^(polarity&1), b^((polarity>>1)&1), c^((polarity>>2)&1))
			case "buffer":
				output = gia.Gia_ManAppendBuf(g, a^(polarity&1))
			}
			gia.Gia_ManAppendCo(g, output)
			gia.Gia_ManAppendCo(g, 0)
			gia.Gia_ManSetRegNum(g, 1)
			m := Int2_ManCreate(g, &Int2_ManPars_t{})
			m.pSatPref = sat_solver_new()
			cos := vec.Vec_IntAlloc(1)
			vec.Vec_IntPush(cos, 0)
			Int2_ManCreateFrames(m, 0, cos)
			out := Int2_ManCheckFrames(m, 0, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
			for mask := int32(0); mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, false}
				want, err := g.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				assumptions := []int32{}
				for i, lit := range []int32{a, b, c} {
					v := Int2_ManCheckFrames(m, 0, lit>>1)
					if v >= 0 {
						assumptions = append(assumptions, v^(1-((mask>>i)&1)))
					}
				}
				for value := int32(0); value < 2; value++ {
					query := append(append([]int32(nil), assumptions...), out^(1-value))
					got := sat_solver_solve(m.pSatPref, query, 0, 0, 0, 0) == l_True
					if got != (want[0] == (value != 0)) {
						t.Fatalf("%s polarity=%d inputs=%d output=%d", kind, polarity, mask, value)
					}
				}
			}
			Int2_ManStop(m)
			gia.Gia_ManStop(g)
		}
	}
}

func TestCreateFramesNativeGatesAcrossRegisters(t *testing.T) {
	g := gia.Gia_ManStart(24)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	state := gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	next := gia.Gia_ManAppendMuxReal(g, a, b^1, state)
	output := gia.Gia_ManAppendXorReal(g, state, a)
	gia.Gia_ManAppendCo(g, output)
	gia.Gia_ManAppendCo(g, next)
	gia.Gia_ManSetRegNum(g, 1)
	m := Int2_ManCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	cos := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cos, 0)
	Int2_ManCreateFrames(m, 0, cos)
	Int2_ManCreateFrames(m, 1, cos)
	out := Int2_ManCheckFrames(m, 1, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
	for mask := int32(0); mask < 16; mask++ {
		a0, b0, a1, b1 := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0
		first, err := g.Evaluate([]bool{a0, b0, false})
		if err != nil {
			t.Fatal(err)
		}
		second, err := g.Evaluate([]bool{a1, b1, first[1]})
		if err != nil {
			t.Fatal(err)
		}
		var assumptions []int32
		for frame := int32(0); frame < 2; frame++ {
			for i, input := range []int32{a, b} {
				lit := Int2_ManCheckFrames(m, frame, input>>1)
				if lit >= 0 {
					assumptions = append(assumptions, lit^(1-((mask>>(2*frame+int32(i)))&1)))
				}
			}
		}
		for value := int32(0); value < 2; value++ {
			query := append(append([]int32(nil), assumptions...), out^(1-value))
			if got := sat_solver_solve(m.pSatPref, query, 0, 0, 0, 0) == l_True; got != (second[0] == (value != 0)) {
				t.Fatalf("mask=%d output=%d expected %v", mask, value, second[0])
			}
		}
	}
}
