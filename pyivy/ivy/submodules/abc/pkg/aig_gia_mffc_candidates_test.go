package abc

// Go-only tests for the deterministic Gia_ManComputeDomsTry replacement.

import (
	"errors"
	"slices"
	"testing"
)

func evaluateGiaCutRoot(t *testing.T, m *GiaManager, root GiaID, boundary []GiaID, assignment int) bool {
	t.Helper()
	leaves := make(map[GiaID]bool, len(boundary))
	for i, id := range boundary {
		leaves[id] = assignment&(1<<i) != 0
	}
	known := map[GiaID]bool{0: false}
	var visit func(GiaID) bool
	visit = func(id GiaID) bool {
		if value, ok := leaves[id]; ok {
			return value
		}
		if value, ok := known[id]; ok {
			return value
		}
		object := m.objects[id]
		read := func(l GiaLit) bool { return visit(l.ID()) != l.Complemented() }
		first, err := object.Fanin0(id)
		if err != nil {
			t.Fatal(err)
		}
		value := read(first)
		if !object.IsBuffer() {
			second, err := object.Fanin1(id)
			if err != nil {
				t.Fatal(err)
			}
			if control, yes := m.MuxControl(id); yes {
				if read(control) {
					value = read(second)
				}
			} else if object.IsXor() {
				value = value != read(second)
			} else {
				value = value && read(second)
			}
		}
		known[id] = value
		return value
	}
	return visit(root)
}

func checkGiaDominatedMffcCandidates(t *testing.T, source *GiaManager, candidates []GiaDominatedMffcCandidate) {
	t.Helper()
	for index, candidate := range candidates {
		if candidate.Cone == nil || candidate.Cone.Name != source.Name || candidate.Cone.Spec != source.Spec || candidate.Cone.InputCount() != len(candidate.Mffc.Boundary) || candidate.Cone.OutputCount() != 1 {
			t.Fatal("candidate metadata", index, candidate)
		}
		for i := 1; i < len(candidate.Mffc.BoundaryReferenceDeltas); i++ {
			if candidate.Mffc.BoundaryReferenceDeltas[i-1] > candidate.Mffc.BoundaryReferenceDeltas[i] {
				t.Fatal("unsorted costs", candidate.Mffc)
			}
		}
		if len(candidate.Mffc.Boundary) > 12 {
			t.Fatal("test graph boundary too large")
		}
		for assignment := 0; assignment < 1<<len(candidate.Mffc.Boundary); assignment++ {
			inputs := make([]bool, len(candidate.Mffc.Boundary))
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			got, err := candidate.Cone.Evaluate(inputs)
			want := evaluateGiaCutRoot(t, source, candidate.Root, candidate.Mffc.Boundary, assignment)
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatal(index, assignment, got, want, err)
			}
		}
	}
}

func TestGiaDominatedMffcCandidates(t *testing.T) {
	g := NewGiaManager()
	g.Name, g.Spec = "candidates", "candidates.aig"
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(n, c)
	r, _ := g.AppendAnd(n, s)
	coR, _ := g.AppendCO(r)
	coN, _ := g.AppendCO(n)
	before := slices.Clone(g.objects)
	candidates, err := g.DominatedMffcCandidates(GiaDominatedMffcOptions{MinimumSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 3 || candidates[0].Object != n.ID() || candidates[0].Root != n.ID() || candidates[1].Object != coR.ID() || candidates[1].Root != r.ID() || candidates[2].Object != coN.ID() || candidates[2].Root != n.ID() {
		t.Fatal("candidate order", candidates)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	checkGiaDominatedMffcCandidates(t, g, candidates)
	if defaultCandidates, err := g.DominatedMffcCandidates(GiaDominatedMffcOptions{}); err != nil || len(defaultCandidates) != 0 {
		t.Fatal("source minimum", defaultCandidates, err)
	}
	oldBoundary := slices.Clone(candidates[1].Mffc.Boundary)
	candidates[1].Mffc.Boundary[0] = 0
	candidates[1].Cone.Name = "changed"
	again, err := g.DominatedMffcCandidates(GiaDominatedMffcOptions{MinimumSize: 1})
	if err != nil || !slices.Equal(again[1].Mffc.Boundary, oldBoundary) || again[1].Cone.Name != g.Name {
		t.Fatal("result aliases", again, err)
	}
}

func TestGiaDominatedMffcCandidateLimits(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	g.AppendCO(n)
	g.AppendCO(n)
	for _, options := range []GiaDominatedMffcOptions{
		{MinimumSize: -1}, {MaxObjects: -1}, {MaxCandidates: -1},
		{MinimumSize: 1, MaxObjects: 1}, {MinimumSize: 1, MaxCandidates: 1},
	} {
		got, err := g.DominatedMffcCandidates(options)
		if got != nil || err == nil {
			t.Fatal(options, got, err)
		}
	}
	if got, err := (*GiaManager)(nil).DominatedMffcCandidates(GiaDominatedMffcOptions{}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}
	for _, threshold := range []int{1, 3} {
		calls := 0
		got, err := g.DominatedMffcCandidates(GiaDominatedMffcOptions{MinimumSize: 1, Stop: func() bool { calls++; return calls >= threshold }})
		if got != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, got, err)
		}
	}
}

func FuzzGiaDominatedMffcCandidates(f *testing.F) {
	f.Add(uint8(3))
	f.Fuzz(func(t *testing.T, mode uint8) {
		g := NewGiaManager()
		g.Name, g.Spec = "fuzz", "fuzz.aig"
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		c, _ := g.AppendCI()
		n, _ := g.AppendAnd(a.NotIf(mode&1 != 0), b)
		s, _ := g.AppendXorReal(n, c.NotIf(mode&2 != 0))
		r, _ := g.AppendMuxReal(a, n, s.NotIf(mode&4 != 0))
		g.AppendCO(r)
		if mode&8 != 0 {
			g.AppendCO(n)
		}
		candidates, err := g.DominatedMffcCandidates(GiaDominatedMffcOptions{MinimumSize: 1, MaxObjects: 128, MaxCandidates: 16})
		if err != nil || len(candidates) == 0 {
			t.Fatal(mode, candidates, err)
		}
		checkGiaDominatedMffcCandidates(t, g, candidates)
	})
}
