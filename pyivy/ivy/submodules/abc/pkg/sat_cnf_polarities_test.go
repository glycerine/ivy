package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestCNFLiteralPolaritiesNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/dar_cuts.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Skip    int
		Ops     [][2]uint32
		Outputs []uint32
		Other   *struct{ Polarities []uint8 }
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, row := range rows {
		if row.Other == nil {
			continue
		}
		count++
		g := darBuildGraph(t, row.Ops, row.Outputs)
		c, e := DeriveAigObjectCNF4(g, AigCNFDeriveOptions{Cuts: DARCutOptions{SkipTruthMinimization: row.Skip != 0}})
		if e != nil {
			t.Fatal(e)
		}
		got, e := c.LiteralPolarities(CNFTransformOptions{})
		if e != nil || !slices.Equal(got, row.Other.Polarities) {
			t.Fatal(got, row.Other.Polarities, e)
		}
	}
	if count != 48 {
		t.Fatal(count)
	}
}
func TestCNFLiteralPolaritiesAggregation(t *testing.T) {
	c := &AigObjectCNF{AigMappedCNF: &AigMappedCNF{variables: 7, clauses: [][]SatLit{{12, 2, 5, 6, 8, 10}, {12, 3, 4}, {13, 2}}}, starts: []int{-1, -1, -1, -1, -1, -1, 0}, counts: []int{-1, -1, -1, -1, -1, -1, 3}}
	got, e := c.LiteralPolarities(CNFTransformOptions{})
	if e != nil || !slices.Equal(got, []uint8{175, 15, 2}) {
		t.Fatal(got, e)
	}
	c.starts[6] = 9
	if r, e := c.LiteralPolarities(CNFTransformOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
	c.starts[6] = 0
	if r, e := c.LiteralPolarities(CNFTransformOptions{Stop: func() bool { return true }}); r != nil || e != ErrCancelled {
		t.Fatal(r, e)
	}
}
func FuzzCNFLiteralPolarities(f *testing.F) {
	f.Add([]byte{2, 5, 6, 8, 10, 3, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		clauses := [][]SatLit{{14}, {15}, {14}}
		for i, b := range data {
			clauses[i%3] = append(clauses[i%3], SatLit(b%14))
			if len(clauses[i%3]) == 17 {
				break
			}
		}
		c := &AigObjectCNF{AigMappedCNF: &AigMappedCNF{variables: 8, clauses: clauses}, starts: []int{-1, -1, -1, -1, -1, -1, -1, 0}, counts: []int{-1, -1, -1, -1, -1, -1, -1, 3}}
		got, e := c.LiteralPolarities(CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for index, clause := range clauses {
			var want uint8
			for slot, l := range clause[1:min(5, len(clause))] {
				var mask uint8
				for _, other := range clauses {
					if other[0]&1 != clause[0]&1 {
						continue
					}
					for _, v := range other[1:] {
						if v.Var() == l.Var() {
							if v&1 != 0 {
								mask |= 1
							} else {
								mask |= 2
							}
						}
					}
				}
				want |= mask << uint(2*slot)
			}
			if got[index] != want {
				t.Fatal(got, index, want)
			}
		}
	})
}
