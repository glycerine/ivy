package sbd

import (
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestTopologyEnumeration(t *testing.T) {
	states := sbdCountTopologies(5, 2)
	// All positive compositions beginning with one and respecting each preceding
	// level's fanin capacity provide an independent exhaustive oracle.
	for size, level := range states {
		var want [][]int32
		var visit func([]int32, int32)
		visit = func(prefix []int32, left int32) {
			if left == 0 {
				want = append(want, append([]int32(nil), prefix...))
				return
			}
			max := int32(1)
			if len(prefix) > 0 {
				max = 2 * prefix[len(prefix)-1]
			}
			for n := int32(1); n <= min(max, left); n++ {
				visit(append(prefix, n), left-n)
			}
		}
		visit(nil, int32(size+1))
		keys := map[string]bool{}
		for _, p := range want {
			keys[sbdTopologyKey(p)] = true
		}
		if len(level) != len(want) {
			t.Fatalf("nodes%d count%d want%d", size+1, len(level), len(want))
		}
		for _, p := range level {
			if !keys[sbdTopologyKey(p)] {
				t.Fatalf("unexpected %v", p)
			}
		}
	}
	if !slices.Equal(states[0][0], []int32{1}) {
		t.Fatal("initial topology")
	}
}
func TestTopologyParameterCount(t *testing.T) {
	set := vec.Vec_IntAlloc(3)
	for _, v := range []int32{1, 2, 3} {
		vec.Vec_IntPush(set, v)
	}
	if got := Sbd_CountConfigVars(set, 4, 2); got != 45 {
		t.Fatalf("count%d", got)
	}
}

func TestTopologyParameterMultiplicationOverflow(t *testing.T) {
	set := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(set, 1<<30)
	defer func() {
		if recover() == nil {
			t.Fatal("overflowed parameter product accepted")
		}
	}()
	Sbd_CountConfigVars(set, 1<<30, 1<<30)
}
