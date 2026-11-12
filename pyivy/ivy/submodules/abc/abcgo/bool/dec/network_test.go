package dec

import "testing"

// A bit-parallel truth network is an independently evaluable concrete adapter.
type truthNetwork struct{}

func (truthNetwork) ConstantOne() any     { return ^uint32(0) }
func (truthNetwork) Complement(v any) any { return ^v.(uint32) }
func (truthNetwork) And(a, b any) any     { return a.(uint32) & b.(uint32) }
func (truthNetwork) Variable(i int) any {
	return []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}[i]
}
func TestNetworkConversion(t *testing.T) {
	g := Dec_GraphCreate(3)
	g.eRoot = Dec_GraphAddNodeMux(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0), Dec_EdgeCreate(2, 0), 0)
	net := truthNetwork{}
	got := Dec_GraphToAig(net, g, []any{net.Variable(0), net.Variable(1), net.Variable(2)}).(uint32)
	if got != Dec_GraphDeriveTruth(g) {
		t.Fatalf("network %08x graph %08x", got, Dec_GraphDeriveTruth(g))
	}
}
func TestFactoring(t *testing.T) {
	for _, sop := range []string{" 0\n", " 1\n", "11 1\n", "1- 1\n-1 1\n", "10 1\n01 1\n", "1-0 1\n-11 1\n", "10 0\n01 0\n"} {
		g := Dec_Factor(sop)
		if Dec_FactorVerify(sop, g) != 1 {
			t.Fatalf("factor failed: %q", sop)
		}
		n := truthNetwork{}
		got := Dec_GraphFactorSop(n, sop).(uint32)
		if got != Dec_GraphDeriveTruth(g) {
			t.Fatalf("SOP network failed: %q", sop)
		}
	}
}

func TestAllThreeVariableTruths(t *testing.T) {
	for truth := 0; truth < 256; truth++ {
		sop := ""
		for a := 0; a < 8; a++ {
			if truth&(1<<a) == 0 {
				continue
			}
			cube := []byte("000")
			for i := range cube {
				if a&(1<<i) != 0 {
					cube[i] = '1'
				}
			}
			sop += string(cube) + " 1\n"
		}
		if sop == "" {
			sop = " 0\n"
		}
		g := Dec_Factor(sop)
		want := uint32(truth) * 0x01010101
		if got := Dec_GraphDeriveTruth(g); got != want {
			t.Fatalf("truth %02x: got %08x wanted %08x", truth, got, want)
		}
	}
}

type missingNetwork struct{ truthNetwork }

func (missingNetwork) LookupAnd(any, any) any { return nil }
func (missingNetwork) SameRegular(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.(uint32) == b.(uint32) || a.(uint32) == ^b.(uint32)
}
func (missingNetwork) Level(any) int           { return 0 }
func (missingNetwork) InDeletionCone(any) bool { return false }
func TestNetworkCountingLimits(t *testing.T) {
	makeGraph := func() *Dec_Graph_t {
		g := Dec_GraphCreate(3)
		g.eRoot = Dec_GraphAddNodeMux(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0), Dec_EdgeCreate(2, 0), 0)
		for i := 0; i < 3; i++ {
			g.SetLeafSignal(i, truthNetwork{}.Variable(i))
		}
		return g
	}
	n := missingNetwork{}
	if got := Dec_GraphToNetworkCount(n, uint32(0), makeGraph(), 3, 2); got != 3 {
		t.Fatal(got)
	}
	if got := Dec_GraphToNetworkCount(n, uint32(0), makeGraph(), 2, 2); got != -1 {
		t.Fatal(got)
	}
	if got := Dec_GraphToNetworkCount(n, uint32(0), makeGraph(), 3, 1); got != -1 {
		t.Fatal(got)
	}
}

func TestMalformedSOPRejected(t *testing.T) {
	for _, sop := range []string{"11 1 extra\n", "11 1\n0\n", "0z 1\n", "11 1\n1 0\n"} {
		t.Run(sop, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("accepted malformed SOP")
				}
			}()
			Dec_Factor(sop)
		})
	}
}
