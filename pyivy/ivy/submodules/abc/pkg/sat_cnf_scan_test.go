package abc

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func cnfScanFixture(t testing.TB, pattern, outputs int) (*AigManager, []AigCNFCut4) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 4)
	for i := range ci {
		ci[i], _ = g.CreateCI()
	}
	a, _ := g.And(ci[0], ci[1])
	b, _ := g.And(ci[2], ci[3])
	root, _ := g.And(a, b)
	g.And(ci[0], ci[2])
	out := root
	if outputs == 3 {
		out = 0
	}
	g.CreateCO(out.NotIf(outputs&1 != 0))
	second := ci[0]
	if outputs == 1 {
		second = a
	}
	if outputs == 2 {
		second = root
	}
	if outputs == 3 {
		second = ci[0].Not()
	}
	g.CreateCO(second)
	cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}, {Root: 6, Leaves: []AigID{3, 4}, Truth: 0x8888}, {Root: 7, Leaves: []AigID{5, 6}, Truth: 0x8888}, {Root: 8, Leaves: []AigID{1, 3}, Truth: 0x8888}}
	switch pattern {
	case 1:
		cuts[2].Leaves = []AigID{1, 2, 3, 4}
		cuts[2].Truth = 0x8000
	case 2:
		cuts[2].Leaves = []AigID{5, 3, 4}
		cuts[2].Truth = 0x8080
	case 3:
		cuts[2].Leaves = nil
		cuts[2].Truth = 65535
	case 4:
		cuts[2].Leaves = []AigID{5}
		cuts[2].Truth = 0xaaaa
	case 5:
		cuts[2].Leaves = []AigID{5, 5}
	case 6:
		cuts[2].Leaves = []AigID{1, 2}
		cuts[2].Truth = 0x6666
	case 7:
		cuts[2].Leaves = []AigID{1, 2, 3, 4}
		cuts[2].Truth = 0x6996
	}
	return g, cuts
}
func TestCNFMappingScanNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_scan.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Pattern, Outputs, Preorder, Area int
		Order                            []AigID
		References                       []uint32
	}
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	if len(cases) != 64 {
		t.Fatal(len(cases))
	}
	for _, c := range cases {
		g, cuts := cnfScanFixture(t, c.Pattern, c.Outputs)
		before := slices.Clone(g.objects)
		r, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: c.Preorder != 0})
		if e != nil {
			t.Fatal(c, e)
		}
		var order []AigID
		for _, cut := range r.Cuts {
			order = append(order, cut.Root)
		}
		if !slices.Equal(order, c.Order) || !slices.Equal(r.References, c.References) || r.Area != c.Area {
			t.Fatal(c, r, order)
		}
		if !reflect.DeepEqual(before, g.objects) {
			t.Fatal("graph changed")
		}
		cnf, e := WriteAigMappedCNF4(g, r.Cuts, 2, MappedCNFOptions{})
		if e != nil {
			t.Fatal(c, e)
		}
		// Arbitrary supplied functions are intentionally used to test scan/writer semantics.
		checkMappedCNFProjection(t, g, r.Cuts, cnf)
	}
}
func TestCNFMappingScanLimitsAndPipeline(t *testing.T) {
	g, cuts := cnfScanFixture(t, 0, 0)
	for _, o := range []CNFMappingScanOptions{{MaxObjects: 1}, {MaxCuts: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		r, e := ScanCNFMapping4(g, cuts, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if r, e := ScanCNFMapping4(g, cuts[:2], CNFMappingScanOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
	duplicated := append(slices.Clone(cuts), cuts[0])
	if r, e := ScanCNFMapping4(g, duplicated, CNFMappingScanOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
	r, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: true})
	if e != nil {
		t.Fatal(e)
	}
	r.Cuts[0].Leaves[0] = 99
	if cuts[2].Leaves[0] != 5 {
		t.Fatal("alias")
	}
	// Area flow should choose the expanded AND4 over two recursive AND2 leaves.
	cuts = append(cuts, AigCNFCut4{Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: 0x8000})
	cnf, e := WriteAigAreaFlowCNF4(g, cuts, 2, MappedCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	expected := []AigCNFCut4{cuts[4]}
	direct, e := WriteAigMappedCNF4(g, expected, 2, MappedCNFOptions{})
	if e != nil || !reflect.DeepEqual(cnf, direct) {
		t.Fatal(cnf, direct, e)
	}
	checkMappedCNFProjection(t, g, expected, cnf)
}

func TestCNFMappingScanSupergates(t *testing.T) {
	for outputs := 0; outputs < 4; outputs++ {
		for _, preorder := range []bool{false, true} {
			g, cuts := cnfScanFixture(t, 0, outputs)
			before := slices.Clone(g.objects)
			got, err := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: preorder, Supergates: []AigID{7}})
			if err != nil {
				t.Fatal(outputs, preorder, err)
			}
			independent, err := ScanAigMapping4(g, cuts, []AigID{7}, AigMappingScanOptions{})
			if err != nil {
				t.Fatal(outputs, preorder, err)
			}
			if got.Area != independent.Area || !slices.Equal(got.References, independent.References) {
				t.Fatal(outputs, preorder, got, independent)
			}
			if slices.Contains(got.Roots, AigID(7)) != slices.Contains(independent.Roots, AigID(7)) {
				t.Fatal("supergate reachability", got.Roots, independent.Roots)
			}
			for _, cut := range got.Cuts {
				if cut.Root == 7 {
					t.Fatal("supergate emitted as ordinary cut", got.Cuts)
				}
			}
			if !reflect.DeepEqual(before, g.objects) {
				t.Fatal("graph changed")
			}
		}
	}
	g, cuts := cnfScanFixture(t, 0, 0)
	for _, supergates := range [][]AigID{{7, 7}, {1}, {99}} {
		if result, err := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Supergates: supergates}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(supergates, result, err)
		}
	}
	if result, err := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Supergates: []AigID{7}, MaxVisits: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
}
func FuzzCNFMappingScan(f *testing.F) {
	f.Add(uint8(0), uint8(0), true)
	f.Fuzz(func(t *testing.T, pattern, outputs uint8, pre bool) {
		g, cuts := cnfScanFixture(t, int(pattern%8), int(outputs%4))
		r, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: pre})
		if e != nil {
			t.Fatal(e)
		}
		refs := make([]uint32, len(g.objects))
		var order []AigID
		area := 0
		var visit func(AigID)
		visit = func(id AigID) {
			refs[id]++
			if refs[id] > 1 || g.objects[id].Type != AigAND {
				return
			}
			c := cuts[int(id)-5]
			if pre {
				order = append(order, id)
			}
			p, _ := CNFSOP4Cover(c.Truth)
			n, _ := CNFSOP4Cover(^c.Truth)
			area += len(p) + len(n)
			for _, leaf := range c.Leaves {
				visit(leaf)
			}
			if !pre {
				order = append(order, id)
			}
		}
		for _, co := range g.cos {
			visit(g.objects[co].Fanin0.ID())
		}
		var actual []AigID
		for _, c := range r.Cuts {
			actual = append(actual, c.Root)
		}
		if area != r.Area || !slices.Equal(refs, r.References) || !slices.Equal(actual, order) {
			t.Fatal(r, area, refs, order)
		}
	})
}

func TestCNFMappingScanDeep(t *testing.T) {
	g := NewAigManager()
	first, _ := g.CreateCI()
	root := first
	var cuts []AigCNFCut4
	for i := 0; i < 10000; i++ {
		input, _ := g.CreateCI()
		next, e := g.And(root, input)
		if e != nil {
			t.Fatal(e)
		}
		cuts = append(cuts, AigCNFCut4{Root: next.ID(), Leaves: []AigID{root.ID(), input.ID()}, Truth: 0x8888})
		root = next
	}
	g.CreateCO(root)
	for _, pre := range []bool{false, true} {
		r, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: pre})
		if e != nil || len(r.Cuts) != len(cuts) || r.Area != 30000 {
			t.Fatal(e)
		}
		want := cuts[0].Root
		if pre {
			want = root.ID()
		}
		if r.Cuts[0].Root != want {
			t.Fatal(r.Cuts[0])
		}
	}
}
