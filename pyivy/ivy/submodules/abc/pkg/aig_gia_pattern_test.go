package abc

import (
	"errors"
	"slices"
	"testing"
)

func giaPatternGraph(t testing.TB, kind int) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a := giaTestCI(t, g)
	b := giaTestCI(t, g)
	c := giaTestCI(t, g)
	var root GiaLit
	var err error
	switch kind {
	case 0:
		root, err = g.AppendAnd(a.Not(), b)
	case 1:
		root, err = g.AppendXorReal(a, b.Not())
	case 2:
		root, err = g.AppendMuxReal(c.Not(), a, b)
	case 3:
		root, err = g.AppendBuffer(a.Not())
	default:
		t.Fatal("bad kind")
	}
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, root.Not())
	return g
}

func TestGiaPatternVerificationExhaustive(t *testing.T) {
	for kind := 0; kind < 4; kind++ {
		g := giaPatternGraph(t, kind)
		before := slices.Clone(g.objects)
		for mask := 0; mask < 8; mask++ {
			for bits := 0; bits < 8; bits++ {
				assignments := make([]GiaPatternAssignment, 0, 3)
				inputs := []GiaTernary{GiaTernaryUnknown, GiaTernaryUnknown, GiaTernaryUnknown}
				for input := 0; input < 3; input++ {
					if mask&(1<<input) != 0 {
						value := bits&(1<<input) != 0
						assignments = append(assignments, GiaPatternAssignment{Input: input, Value: value})
						inputs[input] = GiaTernaryZero
						if value {
							inputs[input] = GiaTernaryOne
						}
					}
				}
				got, err := g.VerifyOutputPattern(0, assignments, GiaCofactorOptions{})
				want, wantErr := g.EvaluateTernary(inputs, GiaCofactorOptions{})
				if err != nil || wantErr != nil || got.Value != want[0] || got.Asserted != (want[0] == GiaTernaryOne) {
					t.Fatal(kind, mask, bits, got, want, err, wantErr)
				}
				if len(got.Cone) == 0 || got.Cone[len(got.Cone)-1] != g.cos[0]-1 {
					t.Fatal("postorder root", kind, got.Cone)
				}
			}
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("source mutation", kind)
		}
	}
}

func TestGiaPatternConeOrderDuplicatesAndErrors(t *testing.T) {
	g := NewGiaManager()
	a := giaTestCI(t, g)
	b := giaTestCI(t, g)
	left, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	root, err := g.AppendAnd(left, a.Not())
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, root)
	cone, err := g.CollectPatternCone(root.ID(), 0)
	if err != nil || !slices.Equal(cone, []GiaID{a.ID(), b.ID(), left.ID(), root.ID()}) {
		t.Fatal(cone, err)
	}
	result, err := g.VerifyOutputPattern(0, []GiaPatternAssignment{{Input: 0, Value: true}, {Input: 0, Value: false}, {Input: 1, Value: true}}, GiaCofactorOptions{})
	if err != nil || result.Value != GiaTernaryZero || result.Asserted {
		t.Fatal(result, err)
	}
	for _, root := range []GiaID{0, g.cos[0], GiaID(g.ObjectCount())} {
		if _, err = g.CollectPatternCone(root, 0); err == nil {
			t.Fatal(root)
		}
	}
	if _, err = g.CollectPatternCone(root.ID(), g.ObjectCount()-1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if got, e := g.VerifyOutputPattern(0, nil, options); e == nil || got.Cone != nil {
			t.Fatal(options, got, e)
		}
	}
	if got, err := g.VerifyOutputPattern(1, nil, GiaCofactorOptions{}); !errors.Is(err, ErrDimension) || got.Cone != nil {
		t.Fatal(got, err)
	}
	if got, err := g.VerifyOutputPattern(0, []GiaPatternAssignment{{Input: 2}}, GiaCofactorOptions{}); !errors.Is(err, ErrDimension) || got.Cone != nil {
		t.Fatal(got, err)
	}
	calls := 0
	if got, err := g.VerifyOutputPattern(0, nil, GiaCofactorOptions{Stop: func() bool { calls++; return calls == 3 }}); !errors.Is(err, ErrCancelled) || got.Cone != nil {
		t.Fatal(calls, got, err)
	}
	constant := NewGiaManager()
	giaTestCO(t, constant, GiaConst1)
	if got, err := constant.VerifyOutputPattern(0, nil, GiaCofactorOptions{}); !errors.Is(err, ErrInvalidLiteral) || got.Cone != nil {
		t.Fatal(got, err)
	}
}

func FuzzGiaPatternVerification(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(0))
	f.Add(uint8(2), uint8(7), uint8(5))
	f.Fuzz(func(t *testing.T, kindByte, maskByte, bitsByte uint8) {
		kind := int(kindByte & 3)
		mask := int(maskByte & 7)
		bits := int(bitsByte & 7)
		g := giaPatternGraph(t, kind)
		assignments := make([]GiaPatternAssignment, 0, 3)
		for input := 0; input < 3; input++ {
			if mask&(1<<input) != 0 {
				assignments = append(assignments, GiaPatternAssignment{Input: input, Value: bits&(1<<input) != 0})
			}
		}
		got, err := g.VerifyOutputPattern(0, assignments, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		possible := uint8(0)
		for completion := 0; completion < 8; completion++ {
			if completion&mask != bits&mask {
				continue
			}
			values := []bool{completion&1 != 0, completion&2 != 0, completion&4 != 0}
			out, evalErr := g.Evaluate(values)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			if out[0] {
				possible |= 2
			} else {
				possible |= 1
			}
		}
		// Local ternary simulation may lose reconvergent correlation, so it
		// may return unknown when completions agree, but never the wrong bit.
		if got.Value != GiaTernaryUnknown && uint8(got.Value)&possible == 0 {
			t.Fatal(kind, mask, bits, got, possible)
		}
	})
}
