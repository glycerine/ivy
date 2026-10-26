package abc

import (
	"slices"
	"testing"
)

func TestGiaNetworkBoxRemoval(t *testing.T) {
	m, mapping, timing, edges := buildGiaEdgeBoxTest(t)
	n, err := mapping.Network(GiaLUTAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := n.WithWindow([]GiaID{4}, [][]GiaID{nil}, GiaLUTAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := proposal.EvaluateDelayWithBoxes(m, edges, timing, GiaLUTAnalysisOptions{})
	if err != nil || r.Timing.Maximum != 11 || r.Timing.Delays[4] != 0 {
		t.Fatal(r, err)
	}
	for _, two := range []bool{false, true} {
		r, err := proposal.SelectEdgesWithBoxes(m, timing, two, GiaLUTAnalysisOptions{})
		want := 11
		if two {
			want = 3
		}
		if err != nil || r.Timing.Maximum != want {
			t.Fatal(two, r, err)
		}
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := proposal.EvaluateDelayWithBoxes(m, edges, timing, options); r.Manager != nil || err == nil {
			t.Fatal(r, err)
		}
		if r, err := proposal.SelectEdgesWithBoxes(m, timing, true, options); r.Edges != nil || r.Manager != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	var missing *GiaLUTNetwork
	if _, err := missing.EvaluateDelayWithBoxes(m, edges, timing, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("nil network")
	}
}
func FuzzGiaNetworkBoxTiming(f *testing.F) {
	f.Add(uint8(0), uint16(12), int16(-5))
	f.Add(uint8(255), uint16(3), int16(7))
	f.Fuzz(func(t *testing.T, flags uint8, a uint16, b int16) {
		m, mapping, timing, edges := buildGiaEdgeBoxTest(t)
		timing.InitPIArrival(0, float32(a%80)/4)
		timing.InitPIArrival(1, float32(b)/4)
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		input := []GiaID{1, 2}
		if flags&1 != 0 {
			input = nil
		}
		output := []GiaID{3, 1}
		if flags&2 != 0 {
			output = []GiaID{3, 1, 3}
		}
		proposal, err := n.WithWindow([]GiaID{4, 5}, [][]GiaID{input, output}, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		records := []GiaLUTRecord{{Root: 5, Fanins: output}, {Root: 6, Fanins: []GiaID{1, 2}}}
		if len(input) > 0 {
			records = append(records, GiaLUTRecord{Root: 4, Fanins: input})
		}
		expected, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		before := timSnapshot(timing)
		actual, err := proposal.EvaluateDelayWithBoxes(m, edges, timing, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want, err := edges.EvaluateDelayWithBoxes(m, expected, timing, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(actual.Timing.Delays, want.Timing.Delays) {
			t.Fatal("network evaluation", err)
		}
		selected, err := proposal.SelectEdgesWithBoxes(m, timing, flags&4 != 0, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		expectedSelection, err := expected.SelectEdgesWithBoxes(m, timing, flags&4 != 0, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(selected.Edges.Rows(), expectedSelection.Edges.Rows()) || !slices.Equal(selected.Timing.Delays, expectedSelection.Timing.Delays) {
			t.Fatal("network selection", err)
		}
		if !slices.Equal(before, timSnapshot(timing)) {
			t.Fatal("source manager changed")
		}
	})
}
