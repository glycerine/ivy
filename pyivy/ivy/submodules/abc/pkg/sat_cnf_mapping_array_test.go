package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestCNFMappingArrayNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/dar_cuts.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Seed    int
		Ops     [][2]uint32
		Outputs []uint32
		Mapping []int32
	}
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, row := range cases {
		if row.Mapping == nil {
			continue
		}
		count++
		g := darBuildGraph(t, row.Ops, row.Outputs)
		before := slices.Clone(g.objects)
		got, e := DeriveAigCNFMappingArray4(g, DARCutOptions{}, CNFMappingArrayOptions{})
		if e != nil || !slices.Equal(got, row.Mapping) {
			t.Fatal(row.Seed, got, row.Mapping, e)
		}
		if !slices.Equal(g.objects, before) {
			t.Fatal("mutated graph")
		}
	}
	if count != 24 {
		t.Fatal(count)
	}
}
func TestCNFMappingArrayLayoutAndLimits(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	cuts := []AigCNFCut4{{Root: 7, Truth: 65535}, {Root: 5, Leaves: []AigID{1, 2}, Truth: 0x8888}}
	r, e := WriteCNFMappingArray4(g, cuts, CNFMappingArrayOptions{})
	if e != nil {
		t.Fatal(e)
	}
	n := len(g.objects)
	if len(r) != n+10 || r[7] != int32(n) || r[5] != int32(n+5) || !slices.Equal(r[n:n+5], []int32{-1, -1, -1, -1, -1}) || uint32(r[n+5]) != 0x88888888 || !slices.Equal(r[n+6:], []int32{1, 2, -1, -1}) {
		t.Fatal(r)
	}
	r[n+6] = 99
	if cuts[1].Leaves[0] != 1 {
		t.Fatal("alias")
	}
	for _, o := range []CNFMappingArrayOptions{{MaxObjects: 1}, {MaxEntries: n + 9}, {Stop: func() bool { return true }}} {
		r, e = WriteCNFMappingArray4(g, cuts, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	for _, bad := range [][]AigCNFCut4{{cuts[0], cuts[0]}, {{Root: 1}}, {{Root: 7, Truth: 1}}, {{Root: 7, Leaves: []AigID{7}, Truth: 0xaaaa}}} {
		r, e = WriteCNFMappingArray4(g, bad, CNFMappingArrayOptions{})
		if r != nil || e == nil {
			t.Fatal(r, e)
		}
	}
	empty, e := WriteCNFMappingArray4(g, nil, CNFMappingArrayOptions{})
	if e != nil || len(empty) != n {
		t.Fatal(empty, e)
	}
	for _, v := range empty {
		if v != 0 {
			t.Fatal(empty)
		}
	}
}
func FuzzCNFMappingArray(f *testing.F) {
	f.Add(uint16(0x8000), false)
	f.Fuzz(func(t *testing.T, truth uint16, reverse bool) {
		g := mappedCNFFixture(t, 0, 0)
		cuts := []AigCNFCut4{{Root: 5, Leaves: []AigID{1, 2, 3, 4}, Truth: truth}, {Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: ^truth}}
		// Root 5 follows all four CIs, so both supplied functions are valid shapes.
		if reverse {
			slices.Reverse(cuts)
		}
		r, e := WriteCNFMappingArray4(g, cuts, CNFMappingArrayOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for i, c := range cuts {
			offset := len(g.objects) + i*5
			if r[c.Root] != int32(offset) || uint32(r[offset]) != uint32(c.Truth)*0x10001 {
				t.Fatal(r, c)
			}
			for j, id := range c.Leaves {
				if r[offset+1+j] != int32(id) {
					t.Fatal(r)
				}
			}
		}
	})
}
