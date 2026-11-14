package cut

import "testing"

type cutManagerStartNetwork struct{ nodes []NetworkNode }

func (n cutManagerStartNetwork) CutNodes() []NetworkNode { return n.nodes }

func TestCutManagerStartForNetworkDefersInternalCuts(t *testing.T) {
	network := cutManagerStartNetwork{nodes: []NetworkNode{
		{ID: 0, Input: true},
		{ID: 1, Input: true},
		{ID: 2, Fanins: [2]int32{0, 1}},
	}}
	p := Cut_ManStartForNetwork(network, DefaultParams(3))
	defer Cut_ManStop(p)
	if p.network[2].ID != 2 || p.networkCounts[0] != 1 || p.networkCounts[1] != 1 {
		t.Fatalf("network inventory was not attached: nodes=%v counts=%v", p.network, p.networkCounts)
	}
	if p.newCuts[0] == nil || p.newCuts[1] == nil {
		t.Fatal("input trivial cuts were not installed")
	}
	if p.newCuts[2] != nil {
		t.Fatal("manager eagerly computed an internal cut")
	}
	if got := Abc_NodeGetCutsRecursive(p, network.nodes[2], 0, 0); got == nil {
		t.Fatal("recursive cut request returned no cut list")
	}
}
