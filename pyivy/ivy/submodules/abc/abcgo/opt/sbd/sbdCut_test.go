package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
)

func cutFromMask(mask uint32) Sbd_Cut_t {
	var c Sbd_Cut_t
	for v := int32(0); v < 6; v++ {
		if mask>>v&1 != 0 {
			c.pLeaves[c.nLeaves] = v
			c.nLeaves++
		}
	}
	c.Sign = Sbd_CutGetSign(&c)
	return c
}
func TestCutMergeAgainstSetUnion(t *testing.T) {
	for a := uint32(0); a < 64; a++ {
		for b := uint32(0); b < 64; b++ {
			ca, cb, want := cutFromMask(a), cutFromMask(b), cutFromMask(a|b)
			for limit := int32(1); limit <= 6; limit++ {
				var result Sbd_Cut_t
				ok := Sbd_CutMergeOrder(&ca, &cb, &result, limit)
				if (ok != 0) != (want.nLeaves <= limit) {
					t.Fatal("union capacity")
				}
				if ok != 0 && (result.nLeaves != want.nLeaves || result.Sign != want.Sign || !slices.Equal(result.pLeaves[:result.nLeaves], want.pLeaves[:want.nLeaves])) {
					t.Fatal("ordered union")
				}
			}
			if (Sbd_CutCheck(&ca, &cb) != 0) != (a&b == b) {
				t.Fatal("cut containment")
			}
		}
	}
}
func TestCutEnumerationTruthAndDelay(t *testing.T) {
	graph := aig.NewGraph()
	var edges []aig.Edge
	for i := 0; i < 5; i++ {
		edges = append(edges, graph.NewInput())
	}
	seed := uint32(7)
	for i := 0; i < 40; i++ {
		seed = seed*1664525 + 1013904223
		a := edges[int(seed)%len(edges)]
		seed = seed*1664525 + 1013904223
		b := edges[int(seed)%len(edges)]
		if seed&1 != 0 {
			a = a.Not()
		}
		if seed&2 != 0 {
			b = b.Not()
		}
		edges = append(edges, graph.AND(a, b))
	}
	graph.AddOutput(edges[len(edges)-1])
	network := NewNetwork(graph)
	for _, maxSize := range []int32{4, 6, 8, 10} {
		store := Sbd_StoAlloc(network, nil, 4, maxSize, 100, 1, 0)
		for _, edge := range graph.LiveEdges() {
			Sbd_StoRefObj(store, int32(edge.NodeID()), -1)
		}
		Sbd_StoComputeCutsConst0(store, 0)
		for _, edge := range graph.Inputs() {
			Sbd_StoComputeCutsCi(store, int32(edge.NodeID()), 0, 0)
		}
		for _, edge := range graph.LiveEdges() {
			if graph.Node(edge).Kind != aig.And {
				continue
			}
			id := int32(edge.NodeID())
			delay := Sbd_StoComputeCutsNode(store, id)
			if delay < 1 {
				t.Fatal("gate delay")
			}
			for _, cut := range store.cuts[id] {
				for primary := uint32(0); primary < 32; primary++ {
					inputs := map[int32]bool{}
					for k, edge := range graph.Inputs() {
						inputs[int32(edge.NodeID())] = primary>>k&1 != 0
					}
					var eval func(int32) bool
					eval = func(lit int32) bool {
						if lit>>1 == 0 {
							return lit&1 != 0
						}
						nodeID := lit >> 1
						value, ok := inputs[nodeID]
						if !ok {
							node := network.object(nodeID)
							value = eval(sourceLiteral(node.Fanins[0])) && eval(sourceLiteral(node.Fanins[1]))
							inputs[nodeID] = value
						}
						return value != (lit&1 != 0)
					}
					var assignment uint32
					for k, leaf := range cut.pLeaves[:cut.nLeaves] {
						if eval(2 * leaf) {
							assignment |= 1 << k
						}
					}
					truth := Sbd_CutTruth(store, &cut)
					got := truth[assignment>>6]>>(assignment&63)&1 != 0
					got = got != (cut.iFunc&1 != 0)
					if got != eval(2*id) {
						t.Fatalf("cut truth size%d node%d primary%d support%v", maxSize, id, primary, cut.pLeaves[:cut.nLeaves])
					}
				}
			}
		}
		Sbd_StoFree(store)
	}
}
