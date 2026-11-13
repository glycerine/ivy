package ssc

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Exercise the public transformations on actual native XOR, MUX and buffer
// objects. Most random sweeping fixtures build those operations out of ANDs.
func TestRandomNativeGateTransformations(t *testing.T) {
	rng := rand.New(rand.NewSource(59201))
	for trial := 0; trial < 50; trial++ {
		g := gia.Gia_ManStart(100)
		g.EnableMuxes()
		var nodes []int32
		for i := 0; i < 4; i++ {
			nodes = append(nodes, gia.Gia_ManAppendCi(g))
		}
		for i := 0; i < 20; i++ {
			// Pick distinct objects so native gates satisfy their preconditions.
			p := rng.Perm(len(nodes))
			a, b, c := nodes[p[0]]^int32(rng.Intn(2)), nodes[p[1]]^int32(rng.Intn(2)), nodes[p[2]]^int32(rng.Intn(2))
			var n int32
			switch i % 4 {
			case 0:
				n = gia.Gia_ManAppendAnd(g, a, b)
			case 1:
				n = gia.Gia_ManAppendXorReal(g, a, b)
			case 2:
				n = gia.Gia_ManAppendMuxReal(g, a, b, c)
			case 3:
				n = gia.Gia_ManAppendBuf(g, a)
			}
			nodes = append(nodes, n)
		}
		for i := 0; i < 4; i++ {
			gia.Gia_ManAppendCo(g, nodes[4+rng.Intn(len(nodes)-4)]^int32(rng.Intn(2)))
		}
		care := gia.Gia_ManStart(10)
		care.EnableMuxes()
		var inputs []int32
		for i := 0; i < 4; i++ {
			inputs = append(inputs, gia.Gia_ManAppendCi(care))
		}
		violation := gia.Gia_ManAppendMuxReal(care, inputs[0], inputs[1], inputs[2]^1)
		gia.Gia_ManAppendCo(care, violation)
		params := DefaultParams()
		params.Verify = true
		params.SatVarMax = 1
		params.CallsRecycle = 0
		result := Ssc_PerformSweeping(g, care, &params)
		checkCare(t, g, result, care)

		for _, transformed := range []*gia.Gia_Man_t{Gia_ManDropContained(g), Gia_ManOptimizeRing(g)} {
			for mask := 0; mask < 16; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				before, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				after, err := transformed.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				union := func(values []bool) bool {
					for _, v := range values {
						if v {
							return true
						}
					}
					return false
				}
				if union(before) != union(after) {
					t.Fatalf("trial %d mask %d: output union changed", trial, mask)
				}
			}
		}
	}
}

func TestDropContainedSequentialCones(t *testing.T) {
	g := gia.Gia_ManStart(32)
	g.EnableMuxes()
	x := gia.Gia_ManAppendCi(g)
	q0 := gia.Gia_ManAppendCi(g)
	q1 := gia.Gia_ManAppendCi(g)
	_ = gia.Gia_ManAppendCi(g) // unused register
	contained := gia.Gia_ManAppendAnd(g, x, q0)
	next0 := gia.Gia_ManAppendXorReal(g, x, q1)
	next1 := gia.Gia_ManAppendMuxReal(g, x, q0, q1)
	gia.Gia_ManAppendCo(g, contained)
	gia.Gia_ManAppendCo(g, q0)
	gia.Gia_ManAppendCo(g, next0)
	gia.Gia_ManAppendCo(g, next1)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 3)

	result := Gia_ManDropContained(g)
	if result.RegisterCount() != 2 || len(result.InputLiterals()) != 3 || len(result.OutputLiterals()) != 3 {
		t.Fatalf("expected one PO and two transitively reachable registers, got %d inputs, %d outputs, %d registers", len(result.InputLiterals()), len(result.OutputLiterals()), result.RegisterCount())
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := result.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want, err := g.Evaluate(append(append([]bool(nil), in...), false))
		if err != nil {
			t.Fatal(err)
		}
		for i := range got {
			if got[i] != want[i+1] {
				t.Fatalf("mask %d output %d: got %v want %v", mask, i, got[i], want[i+1])
			}
		}
	}
}

func TestDropContainedKeepsRegistersWhenNoPrimaryOutputDrops(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x^1)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	result := Gia_ManDropContained(g)
	if result.RegisterCount() != 1 || len(result.OutputLiterals()) != 3 {
		t.Fatalf("all distinct POs and the register must survive: %d outputs, %d registers", len(result.OutputLiterals()), result.RegisterCount())
	}
}
