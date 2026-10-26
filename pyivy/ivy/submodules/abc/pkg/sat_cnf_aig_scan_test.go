package abc

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigScanSupergates(pattern int) []AigID {
	var ids []AigID
	if pattern == 4 || pattern == 6 || pattern == 7 {
		ids = append(ids, 7)
	}
	if pattern == 5 || pattern == 7 {
		ids = append(ids, 5)
	}
	return ids
}

func TestAigMappingScanNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_aig_scan.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Pattern, Outputs, Area int
		Order                  []AigID
		References             []uint32
	}
	if e = json.Unmarshal(data, &rows); e != nil || len(rows) != 32 {
		t.Fatal(len(rows), e)
	}
	for _, row := range rows {
		g, cuts := cnfScanFixture(t, row.Pattern, row.Outputs)
		before := slices.Clone(g.objects)
		r, e := ScanAigMapping4(g, cuts, aigScanSupergates(row.Pattern), AigMappingScanOptions{})
		if e != nil || r.Area != row.Area || !slices.Equal(r.Roots, row.Order) || !slices.Equal(r.References, row.References) {
			t.Fatal(row, r, e)
		}
		if !reflect.DeepEqual(before, g.objects) {
			t.Fatal("graph changed")
		}
	}
}

func TestAigMappingScanLimitsContradictionAndDeep(t *testing.T) {
	g, cuts := cnfScanFixture(t, 0, 0)
	for _, o := range []AigMappingScanOptions{{MaxObjects: 1}, {MaxCuts: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		r, e := ScanAigMapping4(g, cuts, nil, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if r, e := ScanAigMapping4(g, cuts[:2], nil, AigMappingScanOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
	if r, e := ScanAigMapping4(g, cuts, []AigID{7, 7}, AigMappingScanOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
	bad := slices.Clone(cuts)
	bad = append(bad, cuts[0])
	if r, e := ScanAigMapping4(g, bad, nil, AigMappingScanOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}

	contradiction := NewAigManager()
	x, _ := contradiction.CreateCI()
	a, _ := contradiction.CreateCI()
	b, _ := contradiction.CreateCI()
	left, _ := contradiction.And(x, a)
	right, _ := contradiction.And(x.Not(), b)
	root, _ := contradiction.And(left, right)
	contradiction.CreateCO(root)
	r, e := ScanAigMapping4(contradiction, nil, []AigID{root.ID()}, AigMappingScanOptions{})
	if e != nil || r.Area != 1 || !slices.Equal(r.Roots, []AigID{root.ID()}) || r.References[root.ID()] != 1 {
		t.Fatal(r, e)
	}
	for id, n := range r.References {
		if AigID(id) != root.ID() && n != 0 {
			t.Fatal(id, r.References)
		}
	}

	deep := NewAigManager()
	first, _ := deep.CreateCI()
	chain := first
	var deepCuts []AigCNFCut4
	for i := 0; i < 10000; i++ {
		input, _ := deep.CreateCI()
		next, e := deep.And(chain, input)
		if e != nil {
			t.Fatal(e)
		}
		deepCuts = append(deepCuts, AigCNFCut4{Root: next.ID(), Leaves: []AigID{chain.ID(), input.ID()}, Truth: 0x8888})
		chain = next
	}
	deep.CreateCO(chain)
	r, e = ScanAigMapping4(deep, deepCuts, nil, AigMappingScanOptions{})
	if e != nil || len(r.Roots) != 10000 || r.Area != 30000 || r.Roots[0] != chain.ID() {
		t.Fatal(len(r.Roots), r.Area, e)
	}
}

func FuzzAigMappingScan(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(0))
	f.Fuzz(func(t *testing.T, pattern, outputs, mask uint8) {
		g, cuts := cnfScanFixture(t, int(pattern%8), int(outputs%4))
		var supers []AigID
		for i := 0; i < 4; i++ {
			if mask>>uint(i)&1 != 0 {
				supers = append(supers, AigID(5+i))
			}
		}
		r, e := ScanAigMapping4(g, cuts, supers, AigMappingScanOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if len(r.References) != len(g.objects) || r.Area < len(r.Roots) {
			t.Fatal(r)
		}
		seen := make([]bool, len(g.objects))
		for _, id := range r.Roots {
			if seen[id] || g.objects[id].Type != AigAND || r.References[id] == 0 {
				t.Fatal(r)
			}
			seen[id] = true
		}
		for _, co := range g.cos {
			if r.References[g.objects[co].Fanin0.ID()] == 0 {
				t.Fatal(r)
			}
		}
	})
}
