package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func composeTruth32(t testing.TB, n int, words []uint32) *TruthTable {
	t.Helper()
	out := make([]uint64, max(1, (len(words)+1)/2))
	for i, w := range words {
		out[i/2] |= uint64(w) << uint((i%2)*32)
	}
	if n < 6 {
		out[0] |= uint64(words[0]) << 32
	}
	truth, e := TruthTableFromWords(n, out, 0)
	if e != nil {
		t.Fatal(e)
	}
	return truth
}
func TestCNFCompositionCorrectedNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_compose.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		TopTruth, FanTruth, OriginalNativeFirstWord uint32
		TopLeaves, FanLeaves, Leaves                []AigID
		TruthWords                                  []uint32
		Cost                                        int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 240 {
		t.Fatal(len(rows))
	}
	divergent := 0
	for index, row := range rows {
		top := AigCNFCut{Root: 20, Leaves: row.TopLeaves, Truth: composeTruth32(t, len(row.TopLeaves), []uint32{row.TopTruth})}
		fan := AigCNFCut{Root: 10, Leaves: row.FanLeaves, Truth: composeTruth32(t, len(row.FanLeaves), []uint32{row.FanTruth})}
		r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{})
		if e != nil {
			t.Fatal(index, e)
		}
		expected := composeTruth32(t, len(row.Leaves), row.TruthWords)
		if !slices.Equal(r.Cut.Leaves, row.Leaves) || !slices.Equal(r.Cut.Truth.words, expected.words) || r.Cost != row.Cost {
			t.Fatal(index, r.Cut.Leaves, r.Cut.Truth.words, r.Cost, row)
		}
		if row.OriginalNativeFirstWord != row.TruthWords[0] {
			divergent++
		}
		checkCNFComposition(t, top, fan, r)
	}
	if divergent == 0 {
		t.Fatal("native regression lost")
	}
}
func checkCNFComposition(t testing.TB, top, fan AigCNFCut, r *CNFComposition) {
	t.Helper()
	for assignment := 0; assignment < 1<<uint(len(r.Cut.Leaves)); assignment++ {
		value := func(id AigID) int {
			for i, leaf := range r.Cut.Leaves {
				if leaf == id {
					return (assignment >> uint(i)) & 1
				}
			}
			t.Fatal("missing leaf", id)
			return 0
		}
		fanIndex := 0
		for i, id := range fan.Leaves {
			fanIndex |= value(id) << uint(i)
		}
		fanValue := int((fan.Truth.words[fanIndex/64] >> uint(fanIndex%64)) & 1)
		topIndex := 0
		for i, id := range top.Leaves {
			v := fanValue
			if id != fan.Root {
				v = value(id)
			}
			topIndex |= v << uint(i)
		}
		want := (top.Truth.words[topIndex/64] >> uint(topIndex%64)) & 1
		got := (r.Cut.Truth.words[assignment/64] >> uint(assignment%64)) & 1
		if got != want {
			t.Fatal(assignment, r)
		}
	}
}
func TestCNFCompositionLimitsAndRegression(t *testing.T) {
	truth := composeTruth32(t, 2, []uint32{0x88888888})
	top := AigCNFCut{Root: 5, Leaves: []AigID{3, 4}, Truth: truth}
	fan := AigCNFCut{Root: 4, Leaves: []AigID{1, 2}, Truth: truth}
	r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{})
	if e != nil || r.Cut.Truth.words[0] != 0x8080808080808080 || r.Cost != 4 {
		t.Fatal(r, e)
	}
	for _, o := range []CNFCompositionOptions{{MaxInputs: 2}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		r, e = ComposeCNFCuts(top, fan, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	fan.Root = 2
	if r, e = ComposeCNFCuts(top, fan, CNFCompositionOptions{}); r != nil || e == nil {
		t.Fatal(r, e)
	}
}

func TestFormatCNFCut(t *testing.T) {
	cut := AigCNFCut{Root: 20, Leaves: []AigID{1, 7, 12}}
	if got, want := FormatCNFCut(cut), "{1 7 12  } "; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := FormatCNFCut(AigCNFCut{}), "{ } "; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
func FuzzCNFComposition(f *testing.F) {
	f.Add(uint16(0x8888), uint16(0x8888), false)
	f.Fuzz(func(t *testing.T, a, b uint16, overlap bool) {
		topLeaves := []AigID{5, 6, 7, 10}
		if overlap {
			topLeaves = []AigID{1, 2, 3, 10}
		}
		top := AigCNFCut{Root: 20, Leaves: topLeaves, Truth: composeTruth32(t, 4, []uint32{uint32(a) * 0x10001})}
		fan := AigCNFCut{Root: 10, Leaves: []AigID{1, 2, 3, 4}, Truth: composeTruth32(t, 4, []uint32{uint32(b) * 0x10001})}
		r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkCNFComposition(t, top, fan, r)
	})
}

func TestCNFCompositionSixteenInputs(t *testing.T) {
	andTruth := func(n int) *TruthTable {
		words := make([]uint64, 1<<uint(n-6))
		words[len(words)-1] = uint64(1) << 63
		v, e := TruthTableFromWords(n, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	top := AigCNFCut{Root: 40, Leaves: []AigID{17, 18, 19, 20, 21, 22, 23, 24, 30}, Truth: andTruth(9)}
	fan := AigCNFCut{Root: 30, Leaves: []AigID{1, 2, 3, 4, 5, 6, 7, 8}, Truth: andTruth(8)}
	beforeTop, beforeFan := slices.Clone(top.Truth.words), slices.Clone(fan.Truth.words)
	r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{MaxInputs: 16})
	if e != nil || r.Cost != 17 || len(r.Cut.Leaves) != 16 {
		t.Fatal(r, e)
	}
	checkCNFComposition(t, top, fan, r)
	if !slices.Equal(top.Truth.words, beforeTop) || !slices.Equal(fan.Truth.words, beforeFan) {
		t.Fatal("changed source")
	}
	if r, e = ComposeCNFCuts(top, fan, CNFCompositionOptions{}); r != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	if r, e = ComposeCNFCuts(top, fan, CNFCompositionOptions{MaxInputs: 16, ISOP: TruthISOPOptions{MaxElements: 1}}); r != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
}

func TestCNFCompositionMalformedAndCancellation(t *testing.T) {
	truth := composeTruth32(t, 2, []uint32{0x88888888})
	top := AigCNFCut{Root: 5, Leaves: []AigID{3, 4}, Truth: truth}
	fan := AigCNFCut{Root: 4, Leaves: []AigID{1, 2}, Truth: truth}
	for _, bad := range []AigCNFCut{{Root: 4, Leaves: []AigID{2, 1}, Truth: truth}, {Root: 4, Leaves: []AigID{1, 1}, Truth: truth}, {Root: 4, Leaves: []AigID{1, 4}, Truth: truth}, {Root: 4, Leaves: []AigID{1}, Truth: truth}, {Root: 4, Leaves: []AigID{1, 2}}} {
		if r, e := ComposeCNFCuts(top, bad, CNFCompositionOptions{}); r != nil || e == nil {
			t.Fatal(r, e)
		}
	}
	calls := 0
	r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{Stop: func() bool { calls++; return calls > 10 }})
	if r != nil || e != ErrCancelled {
		t.Fatal(r, e)
	}
}

func TestCNFCompositionSaturatedCost(t *testing.T) {
	parity := composeTruth32(t, 4, []uint32{0x69966996})
	top := AigCNFCut{Root: 20, Leaves: []AigID{5, 6, 7, 10}, Truth: parity}
	fan := AigCNFCut{Root: 10, Leaves: []AigID{1, 2, 3, 4}, Truth: parity}
	r, e := ComposeCNFCuts(top, fan, CNFCompositionOptions{})
	if e != nil || r.Cost != 127 {
		t.Fatal(r, e)
	}
	checkCNFComposition(t, top, fan, r)
}
