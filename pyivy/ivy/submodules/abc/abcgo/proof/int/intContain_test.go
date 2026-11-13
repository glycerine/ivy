package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func unaryGraph(invert bool) *aig.Graph {
	g := aig.NewGraph()
	x := g.NewInput()
	if invert {
		x = x.Not()
	}
	g.AddOutput(x)
	return g
}

func TestContainmentAndEquivalence(t *testing.T) {
	x, notX := unaryGraph(false), unaryGraph(true)
	zero := aig.NewGraph()
	zero.NewInput()
	zero.AddOutput(zero.Const0())
	one := aig.NewGraph()
	one.NewInput()
	one.AddOutput(one.Const1())
	for _, test := range []struct {
		name                    string
		a, b                    *aig.Graph
		containment, equivalent int32
	}{
		{"same", x, x, 1, 1},
		{"complements", x, notX, 0, 0},
		{"empty subset", zero, x, 1, 0},
		{"nonempty not subset", one, x, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := Inter_ManCheckContainment(test.a, test.b); got != test.containment {
				t.Fatalf("containment %d want %d", got, test.containment)
			}
			if got := Inter_ManCheckEquivalence(test.a, test.b); got != test.equivalent {
				t.Fatalf("equivalence %d want %d", got, test.equivalent)
			}
		})
	}
}

func TestLatchFramesAndInductiveContainment(t *testing.T) {
	for _, toggle := range []bool{false, true} {
		g := aig.NewGraph()
		x, lo := g.NewInput(), g.NewInput()
		g.AddOutput(x)
		next := lo
		if toggle {
			next = next.Not()
		}
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		var mapping []aig.Edge
		frames := Inter_ManFramesLatches(g, 2, &mapping)
		if len(mapping) != 3 || frames.InputCount() != 3 {
			t.Fatalf("frames inputs=%d mapping=%v", frames.InputCount(), mapping)
		}
		for _, e := range mapping {
			frames.AddOutput(e)
		}
		for mask := 0; mask < 8; mask++ {
			initial := mask&1 != 0
			got, err := frames.Evaluate([]bool{initial, mask&2 != 0, mask&4 != 0})
			want1 := initial != toggle
			if err != nil || len(got) != 3 || got[0] != initial || got[1] != want1 || got[2] != initial {
				t.Fatalf("toggle=%v mask=%d got=%v err=%v", toggle, mask, got, err)
			}
		}
		want := int32(1)
		if toggle {
			want = 0
		}
		for _, backwards := range []int32{0, 1} {
			if got := Inter_ManCheckInductiveContainment(g, unaryGraph(false), 1, backwards); got != want {
				t.Fatalf("toggle=%v backwards=%d got=%d want=%d", toggle, backwards, got, want)
			}
		}
	}
}

func TestOriginalUniquenessLimit(t *testing.T) {
	if Inter_ManCheckUniqueness(nil, nil, nil, 1) != 1 {
		t.Fatal("single frame")
	}
	defer func() {
		if recover() == nil {
			t.Error("expected original multi-frame assertion")
		}
	}()
	Inter_ManCheckUniqueness(nil, nil, nil, 2)
}

// intContain.c constrains the first state negatively for backward checks and
// the final state negatively for forward checks. Enumerate every input trace.
func TestInductiveContainmentAgainstTraceEnumeration(t *testing.T) {
	for transition := 0; transition < 4; transition++ {
		trans := aig.NewGraph()
		input, state := trans.NewInput(), trans.NewInput()
		var next aig.Edge
		switch transition {
		case 0:
			next = state
		case 1:
			next = state.Not()
		case 2:
			next = input
		case 3:
			next = trans.XOR(input, state)
		}
		trans.AddOutput(next)
		aig.Aig_ManSetRegNum(trans, 1)
		for predicate := 0; predicate < 4; predicate++ {
			inter := aig.NewGraph()
			x := inter.NewInput()
			var value aig.Edge
			switch predicate {
			case 0:
				value = inter.Const0()
			case 1:
				value = x
			case 2:
				value = x.Not()
			case 3:
				value = inter.Const1()
			}
			inter.AddOutput(value)
			for steps := int32(0); steps <= 4; steps++ {
				for backward := int32(0); backward <= 1; backward++ {
					want := int32(1)
					for trace := 0; trace < 1<<uint(steps+1); trace++ {
						state := trace & 1
						matches := true
						for frame := int32(0); frame <= steps; frame++ {
							v := predicate == 3 || predicate == 1 && state == 1 || predicate == 2 && state == 0
							if backward == 0 && frame == steps || backward == 1 && frame == 0 {
								v = !v
							}
							matches = matches && v
							if frame == steps {
								break
							}
							in := (trace >> uint(frame+1)) & 1
							switch transition {
							case 1:
								state ^= 1
							case 2:
								state = in
							case 3:
								state ^= in
							}
						}
						if matches {
							want = 0
							break
						}
					}
					if got := Inter_ManCheckInductiveContainment(trans, inter, steps, backward); got != want {
						t.Fatalf("transition=%d predicate=%d steps=%d backward=%d got=%d want=%d", transition, predicate, steps, backward, got, want)
					}
				}
			}
		}
	}
}
