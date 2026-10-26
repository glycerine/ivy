package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaCollectTfiNativeAndOwnership(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	s, err := g.AppendAnd(n, c)
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.AppendAnd(n, s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_tfi.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]GiaID
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 4 {
		t.Fatal(err)
	}
	supportRaw, err := os.ReadFile("test_vectors/oracle/gia_support.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeSupport [][]GiaID
	if err = json.Unmarshal(supportRaw, &nativeSupport); err != nil || len(nativeSupport) != 4 {
		t.Fatal(err)
	}
	combinedRaw, err := os.ReadFile("test_vectors/oracle/gia_nodes_cis.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeCombined [][]GiaID
	if err = json.Unmarshal(combinedRaw, &nativeCombined); err != nil || len(nativeCombined) != 4 {
		t.Fatal(err)
	}
	statsRaw, err := os.ReadFile("test_vectors/oracle/gia_cone_stats.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeStats [][2]int
	if err = json.Unmarshal(statsRaw, &nativeStats); err != nil || len(nativeStats) != 4 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for i, roots := range [][]GiaID{{r.ID()}, {s.ID(), n.ID()}, {n.ID(), r.ID(), n.ID()}, {a.ID()}} {
		stats, err := g.ConeStatistics(roots, 0)
		if err != nil || stats.Inputs != nativeStats[i][0] || stats.Gates != nativeStats[i][1] {
			t.Fatal("native cone counts", stats, nativeStats[i], err)
		}
		supportSize, err := g.SupportSize(roots, 0)
		if err != nil || supportSize != stats.Inputs {
			t.Fatal(supportSize, err)
		}
		coneSize, err := g.ConeSize(roots, 0)
		if err != nil || coneSize != stats.Gates {
			t.Fatal(coneSize, err)
		}
		combined, err := g.CollectNodesAndInputs(roots, 0)
		if err != nil || !slices.Equal(combined, nativeCombined[i]) {
			t.Fatal("native gate/CI postorder", combined, nativeCombined[i], err)
		}
		support, err := g.CollectInputs(roots, 0)
		if err != nil || !slices.Equal(support, nativeSupport[i]) {
			t.Fatal("native support", support, nativeSupport[i], err)
		}
		order, err := g.CollectTransitiveFanins(roots, 0)
		if err != nil || !slices.Equal(order, native[i]) {
			t.Fatal(roots, order, native[i], err)
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	for _, roots := range [][]GiaID{nil, {0}, {a.ID(), 0}} {
		order, err := g.CollectTransitiveFanins(roots, 0)
		if err != nil || len(order) != 0 {
			t.Fatal(order, err)
		}
	}
	if _, err := g.CollectTransitiveFanins([]GiaID{GiaID(g.ObjectCount())}, 0); err == nil {
		t.Fatal("invalid root")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.CollectTransitiveFanins([]GiaID{r.ID()}, cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
func TestGiaCollectTfiExtendedAndDeep(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	control, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.AppendMuxReal(control, a, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	co, err := g.AppendCO(buffer)
	if err != nil {
		t.Fatal(err)
	}
	order, err := g.CollectTransitiveFanins([]GiaID{co.ID()}, 0)
	if err != nil || !slices.Equal(order, []GiaID{control.ID(), mux.ID(), buffer.ID()}) {
		t.Fatal("extended fanins", order, err)
	}
	chain := a
	for i := 0; i < 10000; i++ {
		chain, err = g.AppendAnd(chain, b)
		if err != nil {
			t.Fatal(err)
		}
	}
	order, err = g.CollectTransitiveFanins([]GiaID{chain.ID()}, 0)
	if err != nil || len(order) != 10000 || order[len(order)-1] != chain.ID() {
		t.Fatal("deep traversal", len(order), err)
	}
}
