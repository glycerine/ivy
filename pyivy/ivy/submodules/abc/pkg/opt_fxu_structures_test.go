package abc

// Tests for the Go port of the FXU sparse/container kernel.

import (
	"math/rand"
	"slices"
	"sort"
	"testing"
)

func fxuTestBuild(t *testing.T) *FXUMatrix {
	t.Helper()
	m := NewFXUMatrix(FXUOptions{})
	for range 6 {
		if _, err := m.AddVariable(); err != nil {
			t.Fatal(err)
		}
	}
	rows := [][]int{{0, 1, 3}, {0, 1, 4}, {1, 2, 3}, {1, 2, 5}}
	for i, row := range rows {
		cube, err := m.AddCube(5, i)
		if err != nil {
			t.Fatal(err)
		}
		for _, variable := range row {
			if err = m.AddLiteral(cube, variable); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestFXUSparseMatrixAndCoincidence(t *testing.T) {
	m := fxuTestBuild(t)
	if m.EntryCount() != 12 {
		t.Fatalf("entries=%d", m.EntryCount())
	}
	for a := 0; a < 5; a++ {
		for b := a + 1; b < 6; b++ {
			got, err := m.Coincidence(a, b)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			for _, cube := range m.cubes {
				i, j := sort.SearchInts(cube.Literals, a), sort.SearchInts(cube.Literals, b)
				if i < len(cube.Literals) && cube.Literals[i] == a && j < len(cube.Literals) && cube.Literals[j] == b {
					want++
				}
			}
			if got != want {
				t.Fatalf("coincidence %d,%d=%d want %d", a, b, got, want)
			}
		}
	}
	if err := m.DeleteLiteral(0, 3); err != nil {
		t.Fatal(err)
	}
	if got, _ := m.Coincidence(1, 3); got != 1 {
		t.Fatalf("after delete=%d", got)
	}
	if err := m.AddLiteral(0, 3); err != nil {
		t.Fatal(err)
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	vars := m.Variables()
	vars[0].Occurrences[0] = 999
	if err := m.Validate(); err != nil {
		t.Fatal("snapshot mutated matrix")
	}
}

func TestFXUSinglesHeapAndSelection(t *testing.T) {
	m := fxuTestBuild(t)
	if err := m.ComputeSingles(true, 100); err != nil {
		t.Fatal(err)
	}
	singles := m.Singles()
	if len(singles) == 0 {
		t.Fatal("missing singles")
	}
	for i := 1; i < len(singles); i++ {
		if singles[i-1].Weight < singles[i].Weight {
			t.Fatal("singles not ranked")
		}
	}
	selection, err := m.Select()
	if err != nil {
		t.Fatal(err)
	}
	best := -1
	for _, s := range singles {
		if s.Weight > best {
			best = s.Weight
		}
	}
	if selection.Weight != best || selection.SingleID < 0 {
		t.Fatalf("selection=%+v best=%d", selection, best)
	}
	if err = m.Validate(); err != nil {
		t.Fatal(err)
	}
	if err = m.ComputeSingles(false, 2); err != nil {
		t.Fatal(err)
	}
	if len(m.Singles()) > 2 {
		t.Fatal("single cap ignored")
	}
}

func TestFXUDoubleCanonicalizationAndWeight(t *testing.T) {
	m := fxuTestBuild(t)
	id, err := m.AddDouble(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	c4, _ := m.AddCube(5, 4)
	c5, _ := m.AddCube(5, 5)
	for _, v := range []int{2, 3} {
		_ = m.AddLiteral(c4, v)
	}
	for _, v := range []int{2, 4} {
		_ = m.AddLiteral(c5, v)
	}
	id2, err := m.AddDouble(c4, c5)
	if err != nil {
		t.Fatal(err)
	}
	if id != id2 {
		t.Fatalf("expected common divisor IDs %d %d", id, id2)
	}
	d := m.Doubles()[id]
	if !slices.Equal(d.Left, []int{3}) || !slices.Equal(d.Right, []int{4}) || len(d.Pairs) != 2 {
		t.Fatalf("double=%+v", d)
	}
	// Initial -(1+1), followed by bases of size two and one.
	if d.Weight != 3 {
		t.Fatalf("weight=%d", d.Weight)
	}
	if d.Key != FXUPairHash(d.Left, d.Right) {
		t.Fatal("hash mismatch")
	}
	selection, err := m.Select()
	if err != nil {
		t.Fatal(err)
	}
	if selection.DoubleID != id || selection.Weight != 3 {
		t.Fatalf("selection=%+v", selection)
	}
	if err = m.DeleteDoublePair(id, 0, 1); err != nil {
		t.Fatal(err)
	}
	if got := m.Doubles()[id].Weight; got != 0 {
		t.Fatalf("after delete=%d", got)
	}
	if err = m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFXULimitsCancellationAndAtomicity(t *testing.T) {
	stop := false
	m := NewFXUMatrix(FXUOptions{MaxVariables: 2, MaxCubes: 2, MaxEntries: 2, MaxDivisors: 1, MaxOperations: 100, Stop: func() bool { return stop }})
	v0, _ := m.AddVariable()
	v1, _ := m.AddVariable()
	beforeRejected := m.operations
	if _, err := m.AddVariable(); err != ErrLimit {
		t.Fatalf("var limit=%v", err)
	}
	if m.operations != beforeRejected {
		t.Fatal("rejected variable retained operation accounting")
	}
	c0, _ := m.AddCube(v1, 0)
	c1, _ := m.AddCube(v1, 1)
	if err := m.AddLiteral(c0, v0); err != nil {
		t.Fatal(err)
	}
	if err := m.AddLiteral(c1, v1); err != nil {
		t.Fatal(err)
	}
	beforeRejected = m.operations
	if err := m.AddLiteral(c1, v1); err != ErrLimit {
		t.Fatalf("entry limit=%v", err)
	}
	if m.operations != beforeRejected {
		t.Fatal("rejected literal retained operation accounting")
	}
	beforeCubes := m.Cubes()
	beforeVars := m.Variables()
	beforeOps := m.operations
	stop = true
	if err := m.ComputeSingles(true, 10); err != ErrCancelled {
		t.Fatalf("cancel=%v", err)
	}
	if !slices.EqualFunc(beforeCubes, m.Cubes(), func(a, b FXUCube) bool {
		return a.ID == b.ID && a.Output == b.Output && a.Local == b.Local && slices.Equal(a.Literals, b.Literals)
	}) || !slices.EqualFunc(beforeVars, m.Variables(), func(a, b FXUVariable) bool { return a.ID == b.ID && slices.Equal(a.Occurrences, b.Occurrences) }) {
		t.Fatal("cancel mutated matrix")
	}
	if m.operations != beforeOps {
		t.Fatalf("cancel operations=%d want %d", m.operations, beforeOps)
	}
	stop = false
	m.Close()
	if _, err := m.AddVariable(); err != ErrInvariant {
		t.Fatalf("closed=%v", err)
	}
}

func TestFXUDerivedStateInvalidationAndPairPreconditions(t *testing.T) {
	m := fxuTestBuild(t)
	if _, err := m.AddDouble(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.ComputeSingles(true, 100); err != nil {
		t.Fatal(err)
	}
	if len(m.Singles()) == 0 || len(m.Doubles()[0].Pairs) != 1 {
		t.Fatal("missing derived state")
	}
	if err := m.DeleteLiteral(0, 3); err != nil {
		t.Fatal(err)
	}
	if len(m.Singles()) != 0 || len(m.Doubles()[0].Pairs) != 0 || m.Doubles()[0].HeapPos != -1 {
		t.Fatal("literal edit retained stale divisors")
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}

	contained := NewFXUMatrix(FXUOptions{})
	for range 3 {
		_, _ = contained.AddVariable()
	}
	c0, _ := contained.AddCube(2, 0)
	c1, _ := contained.AddCube(2, 1)
	_ = contained.AddLiteral(c0, 0)
	_ = contained.AddLiteral(c1, 0)
	_ = contained.AddLiteral(c1, 1)
	beforeOps := contained.operations
	if _, err := contained.AddDouble(c0, c1); err != ErrInvariant {
		t.Fatalf("contained pair=%v", err)
	}
	if contained.operations != beforeOps || len(contained.Doubles()) != 0 {
		t.Fatal("rejected pair was not atomic")
	}
}

func TestFXUScanCancellationAndDivisorLimit(t *testing.T) {
	calls := 0
	enabled := false
	m := NewFXUMatrix(FXUOptions{MaxDivisors: 1, Stop: func() bool {
		if !enabled {
			return false
		}
		calls++
		return calls == 2
	}})
	for range 4 {
		_, _ = m.AddVariable()
	}
	for i := range 3 {
		cube, _ := m.AddCube(3, i)
		_ = m.AddLiteral(cube, 0)
		_ = m.AddLiteral(cube, 1)
		_ = m.AddLiteral(cube, 2)
	}
	calls = 0
	enabled = true
	beforeOps := m.operations
	if _, err := m.Coincidence(0, 1); err != ErrCancelled {
		t.Fatalf("coincidence cancellation=%v", err)
	}
	if m.operations != beforeOps {
		t.Fatal("cancelled scan retained operation accounting")
	}
	enabled = false
	m.options.Stop = nil
	if err := m.ComputeSingles(true, 100); err != nil {
		t.Fatal(err)
	}
	if len(m.Singles()) != 1 {
		t.Fatalf("divisor limit retained %d singles", len(m.Singles()))
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

type fxuTestModel struct {
	vars    int
	cubes   []map[int]bool
	outputs []int
	locals  []int
}

func fxuTestCheckModel(t *testing.T, m *FXUMatrix, model fxuTestModel) {
	t.Helper()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(m.variables) != model.vars || len(m.cubes) != len(model.cubes) {
		t.Fatalf("dimensions")
	}
	entries := 0
	for cube, set := range model.cubes {
		got := m.cubes[cube].Literals
		want := make([]int, 0, len(set))
		for v := range set {
			want = append(want, v)
		}
		sort.Ints(want)
		if !slices.Equal(got, want) {
			t.Fatalf("cube %d got %v want %v", cube, got, want)
		}
		entries += len(want)
	}
	if entries != m.entries {
		t.Fatalf("entries=%d want=%d", m.entries, entries)
	}
	for a := 0; a < model.vars; a++ {
		for b := a + 1; b < model.vars; b++ {
			want := 0
			for _, set := range model.cubes {
				if set[a] && set[b] {
					want++
				}
			}
			if got := intersectCount(m.variables[a].Occurrences, m.variables[b].Occurrences); got != want {
				t.Fatalf("coin %d,%d=%d want %d", a, b, got, want)
			}
		}
	}
}

func fxuTestRandomCampaign(t *testing.T, seed int64, steps int) {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	m := NewFXUMatrix(FXUOptions{MaxOperations: 1 << 24})
	model := fxuTestModel{}
	for step := 0; step < steps; step++ {
		switch rng.Intn(6) {
		case 0:
			if model.vars < 12 {
				if _, err := m.AddVariable(); err != nil {
					t.Fatal(err)
				}
				model.vars++
			}
		case 1:
			if model.vars > 0 && len(model.cubes) < 30 {
				out := rng.Intn(model.vars)
				local := len(model.cubes)
				id, err := m.AddCube(out, local)
				if err != nil {
					t.Fatal(err)
				}
				if id != len(model.cubes) {
					t.Fatal("cube ID")
				}
				model.cubes = append(model.cubes, map[int]bool{})
				model.outputs = append(model.outputs, out)
				model.locals = append(model.locals, local)
			}
		case 2:
			if model.vars > 0 && len(model.cubes) > 0 {
				c := rng.Intn(len(model.cubes))
				v := rng.Intn(model.vars)
				if !model.cubes[c][v] {
					if err := m.AddLiteral(c, v); err != nil {
						t.Fatal(err)
					}
					model.cubes[c][v] = true
				}
			}
		case 3:
			if len(model.cubes) > 0 {
				c := rng.Intn(len(model.cubes))
				if len(model.cubes[c]) > 0 {
					keys := make([]int, 0, len(model.cubes[c]))
					for v := range model.cubes[c] {
						keys = append(keys, v)
					}
					v := keys[rng.Intn(len(keys))]
					if err := m.DeleteLiteral(c, v); err != nil {
						t.Fatal(err)
					}
					delete(model.cubes[c], v)
				}
			}
		case 4:
			if err := m.ComputeSingles(rng.Intn(2) == 0, 32); err != nil {
				t.Fatal(err)
			}
		case 5:
			if len(model.cubes) > 1 {
				c1, c2 := rng.Intn(len(model.cubes)), rng.Intn(len(model.cubes))
				if c1 != c2 && model.outputs[c1] == model.outputs[c2] {
					_, _ = m.AddDouble(c1, c2)
				}
			}
		}
		fxuTestCheckModel(t, m, model)
	}
}

func TestFXUIndependentStateModel5000(t *testing.T) {
	for i := 0; i < 100; i++ {
		fxuTestRandomCampaign(t, int64(i+1), 50)
	}
}

func FuzzFXUSparseKernel(f *testing.F) {
	f.Add([]byte{0, 0, 1, 2, 3, 4, 5, 2, 3, 4, 1, 5})
	f.Add([]byte{0, 0, 0, 1, 1, 2, 2, 2, 4, 5, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		m := NewFXUMatrix(FXUOptions{MaxVariables: 16, MaxCubes: 32, MaxEntries: 256, MaxDivisors: 128, MaxOperations: 1 << 16})
		model := fxuTestModel{}
		for _, op := range data {
			switch op % 6 {
			case 0:
				if model.vars < 16 {
					_, _ = m.AddVariable()
					model.vars++
				}
			case 1:
				if model.vars > 0 && len(model.cubes) < 32 {
					out := int(op) % model.vars
					local := len(model.cubes)
					id, err := m.AddCube(out, local)
					if err != nil {
						t.Fatal(err)
					}
					if id != len(model.cubes) {
						t.Fatal("id")
					}
					model.cubes = append(model.cubes, map[int]bool{})
					model.outputs = append(model.outputs, out)
					model.locals = append(model.locals, local)
				}
			case 2:
				if model.vars > 0 && len(model.cubes) > 0 {
					c := int(op) % len(model.cubes)
					v := (int(op) / 3) % model.vars
					if !model.cubes[c][v] {
						if err := m.AddLiteral(c, v); err != nil {
							t.Fatal(err)
						}
						model.cubes[c][v] = true
					}
				}
			case 3:
				if len(model.cubes) > 0 {
					c := int(op) % len(model.cubes)
					for v := range model.cubes[c] {
						if err := m.DeleteLiteral(c, v); err != nil {
							t.Fatal(err)
						}
						delete(model.cubes[c], v)
						break
					}
				}
			case 4:
				if err := m.ComputeSingles(op&8 != 0, 16); err != nil {
					t.Fatal(err)
				}
			case 5:
				if len(model.cubes) > 1 {
					c1 := int(op) % len(model.cubes)
					c2 := (int(op)/7 + 1) % len(model.cubes)
					if c1 != c2 && model.outputs[c1] == model.outputs[c2] {
						_, _ = m.AddDouble(c1, c2)
					}
				}
			}
			fxuTestCheckModel(t, m, model)
		}
	})
}
