package abc

import (
	"math/bits"
	"slices"
	"testing"
)

func TestGiaFullWindow(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b.Not())
	y, _ := m.AppendBuffer(x)
	m.AppendCO(y.Not())
	m.AppendCO(x)
	r, e := m.FullWindow(GiaLUTAnalysisOptions{})
	if e != nil || !slices.Equal(r.Leaves, []GiaID{1, 2}) || !slices.Equal(r.Ands, []GiaID{3, 4}) || !slices.Equal(r.Roots, []GiaID{4, 3}) {
		t.Fatal(r, e)
	}
	r.Leaves[0] = 99
	again, e := m.FullWindow(GiaLUTAnalysisOptions{})
	if e != nil || again.Leaves[0] != 1 {
		t.Fatal(again, e)
	}
	empty, e := NewGiaManager().FullWindow(GiaLUTAnalysisOptions{})
	if e != nil || len(empty.Roots)+len(empty.Leaves)+len(empty.Ands) != 0 {
		t.Fatal(empty, e)
	}
}

func TestGiaFullWindowLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxObjects: 3}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if r, e := m.FullWindow(o); e == nil || r.Roots != nil || r.Leaves != nil || r.Ands != nil {
			t.Fatal(r, e)
		}
	}
	if r, e := (*GiaManager)(nil).FullWindow(GiaLUTAnalysisOptions{}); e == nil || r.Roots != nil {
		t.Fatal(r, e)
	}
}

func TestFormatGiaWindowCut(t *testing.T) {
	c := GiaWindowCut{Leaves: [2]uint64{5, 1 << 63}, Nodes: [2]uint64{2, 3}}
	s, n := FormatGiaWindowCut(c)
	if s != "{ i0 i2 i127  n1 n64 n65 };\n" || n != 6 {
		t.Fatal(s, n)
	}
	if s, n = FormatGiaWindowCut(GiaWindowCut{}); s != "{  };\n" || n != 0 {
		t.Fatal(s, n)
	}
}

func TestGiaRemapPassStats(t *testing.T) {
	r := GiaUnitRemapPassResult{
		DiscoveryBuilds: 1, ReferenceRefreshes: 3, SolverBuilds: 2, WindowResets: 4,
		Improvements: 2,
		Attempts: []GiaUnitRemapAttempt{
			{Reason: "improved"}, {Reason: "repeated"}, {Reason: "small"},
			{Reason: "large"}, {Reason: "unchanged"}, {Reason: "improved"},
		},
	}
	s := r.Stats()
	if s != (GiaRemapPassStats{Tried: 6, Improved: 2, Repeated: 1, Small: 1, Large: 1, DiscoveryBuilds: 1, ReferenceRefreshes: 3, SolverBuilds: 2, WindowResets: 4}) {
		t.Fatal(s)
	}
	if text := s.Format(); text != "Tried=6 Improved=2 Repeated=1 Small=1 Large=1 DiscoveryBuilds=1 ReferenceRefreshes=3 SolverBuilds=2 WindowResets=4\n" {
		t.Fatal(text)
	}
}

func FuzzGiaWindowDiagnostic(f *testing.F) {
	f.Add(uint64(5), uint64(1)<<63, uint64(2), uint64(3))
	f.Fuzz(func(t *testing.T, i0, i1, n0, n1 uint64) {
		c := GiaWindowCut{Leaves: [2]uint64{i0, i1}, Nodes: [2]uint64{n0, n1}}
		s, count := FormatGiaWindowCut(c)
		want := bits.OnesCount64(i0) + bits.OnesCount64(i1) + bits.OnesCount64(n0) + bits.OnesCount64(n1)
		if count != want || len(s) < 5 || s[:2] != "{ " || s[len(s)-3:] != "};\n" {
			t.Fatal(count, want, s)
		}
	})
}
