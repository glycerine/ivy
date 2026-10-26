package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func cnfStructuralCuts(g *AigManager) []AigCNFCut4 {
	var cuts []AigCNFCut4
	for id, obj := range g.objects {
		if obj.Type != AigAND {
			continue
		}
		var truth uint16
		for a := 0; a < 16; a++ {
			x := (a&1 != 0) != obj.Fanin0.Complemented()
			y := (a&2 != 0) != obj.Fanin1.Complemented()
			if x && y {
				truth |= 1 << uint(a)
			}
		}
		cuts = append(cuts, AigCNFCut4{Root: AigID(id), Leaves: []AigID{obj.Fanin0.ID(), obj.Fanin1.ID()}, Truth: truth})
	}
	return cuts
}
func TestCNFPostprocessCorrectedNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_post.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Seed, Mode, Changes, Area int
		Ops                       [][2]uint32
		Outputs                   []uint32
		References                []uint32
		Cuts                      []struct {
			Root       AigID
			Leaves     []AigID
			TruthWords []uint32
		}
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 48 {
		t.Fatal(len(rows))
	}
	changes := 0
	for _, row := range rows {
		g := darBuildGraph(t, row.Ops, row.Outputs)
		g.Cleanup()
		before := slices.Clone(g.objects)
		cuts := cnfStructuralCuts(g)
		r, e := PostprocessCNFMapping4(g, cuts, CNFPostprocessOptions{})
		if e != nil {
			t.Fatal(row.Seed, row.Mode, e)
		}
		if r.Changes != row.Changes || r.Area != row.Area || !slices.Equal(r.References, row.References) || len(r.Cuts) != len(row.Cuts) {
			t.Fatal(row.Seed, row.Mode, r, row)
		}
		for i, c := range r.Cuts {
			expected := row.Cuts[i]
			truth := composeTruth32(t, len(expected.Leaves), expected.TruthWords)
			if c.Root != expected.Root || !slices.Equal(c.Leaves, expected.Leaves) || !slices.Equal(c.Truth.words, truth.words) {
				t.Fatal(row.Seed, row.Mode, i, c, expected)
			}
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("mutated graph")
		}
		changes += r.Changes
		cnf, e := WriteAigMappedCNF(g, r.Cuts, 2, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, cnf)
		automatic, e := DeriveAigPostprocessedCNF4(g, 2, AigCNFDeriveOptions{}, CNFPostprocessOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, automatic)
	}
	if changes == 0 {
		t.Fatal("no recovery exercised")
	}
}
func TestCNFPostprocessLimitsAndOwnership(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	cuts := cnfStructuralCuts(g)
	r, e := PostprocessCNFMapping4(g, cuts, CNFPostprocessOptions{})
	if e != nil || r.Changes != 1 || r.Area != 7 || len(r.Cuts) != 2 {
		t.Fatal(r, e)
	}
	r.Cuts[0].Leaves[0] = 99
	if cuts[0].Leaves[0] != 1 {
		t.Fatal("alias")
	}
	for _, o := range []CNFPostprocessOptions{{MaxObjects: 1}, {MaxCuts: 1}, {Stop: func() bool { return true }}} {
		r, e = PostprocessCNFMapping4(g, cuts, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	chain, _ := wideCNFGraph(t, 6, 1, 0)
	if limited, e := PostprocessCNFMapping4(chain, cnfStructuralCuts(chain), CNFPostprocessOptions{MaxAttempts: 1}); limited != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(limited, e)
	}
	// A proposal limit skips recovery rather than discarding the valid mapping.
	r, e = PostprocessCNFMapping4(g, cuts, CNFPostprocessOptions{Composition: CNFCompositionOptions{MaxInputs: 1}})
	if e != nil || r.Changes != 0 || len(r.Cuts) != 3 {
		t.Fatal(r, e)
	}
}
func FuzzCNFPostprocess(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		g.Cleanup()
		cuts := cnfStructuralCuts(g)
		initial, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := PostprocessCNFMapping4(g, cuts, CNFPostprocessOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Area > initial.Area {
			t.Fatal(initial.Area, r.Area)
		}
		cnf, e := WriteAigMappedCNF(g, r.Cuts, 2, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, cnf)
	})
}
