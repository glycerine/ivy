package abc

import (
	"encoding/binary"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestCNFAreaFlowNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_area_flow.bin")
	if e != nil {
		t.Fatal(e)
	}
	if len(data) != 65536*3 {
		t.Fatal(len(data))
	}
	g := mappedCNFFixture(t, 0, 0)
	cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}, {Root: 6, Leaves: []AigID{3, 4}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{1, 2, 3, 4}}}
	for truth := 0; truth < 65536; truth++ {
		cuts[2].Truth = uint16(truth)
		r, e := SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{})
		if e != nil {
			t.Fatal(truth, e)
		}
		if r.Flows[7] != uint32(binary.LittleEndian.Uint16(data[truth*3:])) || r.Values[2] != uint16(data[truth*3+2]) {
			t.Fatal(truth, r)
		}
	}
}
func TestCNFAreaFlowSelection(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}, {Root: 6, Leaves: []AigID{3, 4}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{5, 6}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: 0x8000}}
	r, e := SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{})
	if e != nil {
		t.Fatal(e)
	}
	// Two AND2 cuts cost 30 each; root's recursive flow costs 90, vs AND4's 50.
	if r.Flows[5] != 30 || r.Flows[6] != 30 || r.Flows[7] != 50 || len(r.Cuts[2].Leaves) != 4 {
		t.Fatal(r)
	}
	r.Cuts[0].Leaves[0] = 99
	if cuts[0].Leaves[0] != 1 {
		t.Fatal("alias")
	}
	// Equal clause cost favors the larger reference sum, then first candidate.
	cuts[2] = AigCNFCut4{Root: 7, Leaves: []AigID{2}, Truth: 0xaaaa}
	cuts[3] = AigCNFCut4{Root: 7, Leaves: []AigID{1}, Truth: 0xaaaa}
	r, e = SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{})
	if e != nil || r.Cuts[2].Leaves[0] != 1 {
		t.Fatal(r, e)
	}
	for _, o := range []CNFAreaFlowOptions{{MaxObjects: 1}, {MaxCuts: 1}, {Stop: func() bool { return true }}} {
		r, e = SelectCNFAreaFlow4(g, cuts, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	cuts[0].Truth = 1
	if r, e = SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
}

func TestCNFSuperAreaFlow(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}, {Root: 6, Leaves: []AigID{3, 4}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{5, 6}, Truth: 0x8888}}
	mapping, err := SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{})
	if err != nil {
		t.Fatal(err)
	}
	super := []AigLit{(AigLit(5) << 1).Not(), AigLit(1) << 1, AigConst1}
	want := uint32(100*(len(super)+1)) + mapping.Flows[5]/g.objects[5].RefCount
	got, err := CNFSuperAreaFlow(g, super, mapping.Flows)
	if err != nil || got != want {
		t.Fatal(got, want, err)
	}
	for _, test := range []struct {
		graph *AigManager
		items []AigLit
		flows []uint32
		want  error
	}{
		{nil, nil, nil, ErrDimension},
		{g, super, mapping.Flows[:2], ErrDimension},
		{g, []AigLit{InvalidAigLit}, mapping.Flows, ErrInvalidLiteral},
		{g, []AigLit{AigLit(g.cos[0]) << 1}, mapping.Flows, ErrInvalidLiteral},
	} {
		if value, err := CNFSuperAreaFlow(test.graph, test.items, test.flows); value != 0 || !errors.Is(err, test.want) {
			t.Fatal(value, err, test.want)
		}
	}
	dangling := NewAigManager()
	a, _ := dangling.CreateCI()
	b, _ := dangling.CreateCI()
	root, _ := dangling.And(a, b)
	if value, err := CNFSuperAreaFlow(dangling, []AigLit{root}, make([]uint32, dangling.ObjectCount())); value != 0 || !errors.Is(err, ErrInvariant) {
		t.Fatal(value, err)
	}
	overflow := slices.Clone(mapping.Flows)
	overflow[5] = ^uint32(0)
	if value, err := CNFSuperAreaFlow(g, []AigLit{AigLit(5) << 1}, overflow); value != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(value, err)
	}
}

func FuzzCNFAreaFlowSelection(f *testing.F) {
	f.Add(uint16(0x8000), uint16(0x8888))
	f.Fuzz(func(t *testing.T, a, b uint16) {
		g := mappedCNFFixture(t, 0, 0)
		cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}, {Root: 6, Leaves: []AigID{3, 4}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: a}, {Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: b}}
		r, e := SelectCNFAreaFlow4(g, cuts, CNFAreaFlowOptions{})
		if e != nil {
			t.Fatal(e)
		}
		cost := func(v uint16) int {
			p, e := CNFSOP4Cover(v)
			if e != nil {
				t.Fatal(e)
			}
			n, e := CNFSOP4Cover(^v)
			if e != nil {
				t.Fatal(e)
			}
			return 10 * (len(p) + len(n))
		}
		want := a
		if cost(b) < cost(a) {
			want = b
		}
		if r.Cuts[2].Truth != want || int(r.Flows[7]) != cost(want) {
			t.Fatal(a, b, r)
		}
	})
}

func FuzzCNFSuperAreaFlow(f *testing.F) {
	f.Add(uint64(0x0706050403020100), uint64(0x1020304050607080))
	f.Fuzz(func(t *testing.T, raw, seed uint64) {
		g := mappedCNFFixture(t, 0, 0)
		flows := make([]uint32, g.ObjectCount())
		for i := range flows {
			flows[i] = uint32(seed>>uint((i%8)*8)) & 0xffff
		}
		count := 1 + int(raw&7)
		super := make([]AigLit, count)
		want := uint64(100 * (count + 1))
		for i := range super {
			id := AigID((raw >> uint(3+4*i)) % 8)
			super[i] = (AigLit(id) << 1).NotIf(raw&(uint64(1)<<uint(60+i%4)) != 0)
			object := g.objects[id]
			if object.Type == AigAND {
				want += uint64(flows[id] / object.RefCount)
			}
		}
		got, err := CNFSuperAreaFlow(g, super, flows)
		if err != nil || uint64(got) != want {
			t.Fatal(super, got, want, err)
		}
	})
}
