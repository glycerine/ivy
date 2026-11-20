package core

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func layeredCutGraph() *aig.Graph {
	const width, levels = 16, 48
	g := aig.NewGraph()
	var layer, next [width]aig.Edge
	for i := range layer {
		layer[i] = g.NewInput()
	}
	for k := 0; k < levels; k++ {
		for i, left := range layer {
			right := layer[(i+1+2*(k%5))%width]
			if (i+k)%3 == 0 {
				left = left.Not()
			}
			if (i+2*k)%5 == 0 {
				right = right.Not()
			}
			next[i] = g.AND(left, right)
		}
		layer = next
	}
	for _, output := range layer {
		g.AddOutput(output)
	}
	return g
}

func TestLayeredCutSlotsMatchNative(t *testing.T) {
	g := layeredCutGraph()
	// From test_vectors/perf_cuts.c, over all 768 gates and ten physical slots.
	for _, tc := range []struct {
		skip bool
		want uint64
	}{{false, 0x1468219d9a6a805f}, {true, 0xd9fb95121132cb23}} {
		slots, err := Dar_ManComputeCutSlots(g, 10, tc.skip)
		if err != nil {
			t.Fatal(err)
		}
		var h uint64 = 1
		mix := func(value int) { h = (h ^ uint64(value)) * 1099511628211 }
		for _, edge := range g.LiveEdges() {
			if g.Node(edge).Kind != aig.And {
				continue
			}
			mix(edge.NodeID())
			for i, s := range slots[edge] {
				if !s.Used {
					continue
				}
				mix(i)
				mix(s.Value)
				mix(int(s.Cut.Truth))
				mix(len(s.Cut.Leaves))
				for _, leaf := range s.Cut.Leaves {
					mix(leaf.NodeID())
				}
			}
		}
		if h != tc.want {
			t.Errorf("skip minimization %v: cut digest %016x, want %016x", tc.skip, h, tc.want)
		}
	}
}

func BenchmarkLayeredCutSlots(b *testing.B) {
	for _, tc := range []struct {
		name string
		skip bool
	}{{"Minimize", false}, {"KeepSupport", true}} {
		b.Run(tc.name, func(b *testing.B) {
			g := layeredCutGraph()
			b.ReportAllocs()
			for b.Loop() {
				if _, err := Dar_ManComputeCutSlots(g, 10, tc.skip); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Compare node-owned enumeration with Dar_ManComputeCuts in perf_cuts.c.
func BenchmarkLayeredCutEnumeration(b *testing.B) {
	for _, tc := range []struct {
		name string
		skip bool
	}{{"Minimize", false}, {"KeepSupport", true}} {
		b.Run(tc.name, func(b *testing.B) {
			g := layeredCutGraph()
			b.ReportAllocs()
			for b.Loop() {
				m, err := NewWholeGraphCutEnumerator(g, 10, tc.skip)
				if err != nil {
					b.Fatal(err)
				}
				for _, edge := range g.LiveEdges() {
					if g.Node(edge).Kind != aig.Output {
						m.slots(edge)
					}
				}
			}
		})
	}
}
