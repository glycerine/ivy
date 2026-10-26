package abc

// Tests for the stand-alone Ivy graph core derived from src/aig/ivy.

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func mustIvy(e IvyEdge, err error) IvyEdge {
	if err != nil {
		panic(err)
	}
	return e
}

func TestIvyCoreBooleanTruthAndHashing(t *testing.T) {
	m, err := NewIvyManager(IvyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a := mustIvy(m.CreatePI())
	b := mustIvy(m.CreatePI())
	c := mustIvy(m.CreatePI())
	and := mustIvy(m.And(a, b))
	if again := mustIvy(m.And(b, a)); again != and {
		t.Fatalf("hash mismatch %d %d", and, again)
	}
	xor := mustIvy(m.Xor(a, b))
	mux := mustIvy(m.Mux(c, a, b))
	maj := mustIvy(m.Majority(a, b, c))
	for _, e := range []IvyEdge{and, xor, mux, maj} {
		if _, err = m.CreatePO(e); err != nil {
			t.Fatal(err)
		}
	}
	for bits := 0; bits < 8; bits++ {
		av := bits&1 != 0
		bv := bits&2 != 0
		cv := bits&4 != 0
		got, e := m.Evaluate([]bool{av, bv, cv}, nil)
		if e != nil {
			t.Fatal(e)
		}
		want := []bool{av && bv, av != bv, map[bool]bool{true: av, false: bv}[cv], (av && bv) || (av && cv) || (bv && cv)}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%03b got %v want %v", bits, got, want)
		}
	}
	if err = m.Check(); err != nil {
		t.Fatal(err)
	}
	converted, err := m.ToAigManager()
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		left, _ := m.Evaluate(in, nil)
		right, e := converted.Evaluate(in)
		if e != nil || !reflect.DeepEqual(left, right) {
			t.Fatalf("conversion %03b: %v %v %v", bits, left, right, e)
		}
	}
}

func TestIvyConstantsMultiMiterAndInit(t *testing.T) {
	m, _ := NewIvyManager(IvyOptions{})
	a := mustIvy(m.CreatePI())
	b := mustIvy(m.CreatePI())
	cases := []struct{ x, y, w IvyEdge }{{a, a, a}, {a, a.Not(), IvyConst0}, {IvyConst1, a, a}, {IvyConst0, a, IvyConst0}}
	for _, tc := range cases {
		g, e := m.And(tc.x, tc.y)
		if e != nil || g != tc.w {
			t.Fatalf("and(%d,%d)=(%d,%v), want %d", tc.x, tc.y, g, e, tc.w)
		}
	}
	all := mustIvy(m.Multi([]IvyEdge{a, b, IvyConst1}, IvyAND))
	diff, err := m.AnyDifference([][2]IvyEdge{{a, a}, {a, b}})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []IvyEdge{all, diff} {
		if _, err = m.CreatePO(e); err != nil {
			t.Fatal(err)
		}
	}
	for bits := 0; bits < 4; bits++ {
		av, bv := bits&1 != 0, bits&2 != 0
		got, e := m.Evaluate([]bool{av, bv}, nil)
		if e != nil {
			t.Fatal(e)
		}
		if got[0] != (av && bv) || got[1] != (av != bv) {
			t.Fatalf("truth %d: %v", bits, got)
		}
	}
	if v, e := IvyInitNotIf(IvyInit0, true); e != nil || v != IvyInit1 {
		t.Fatal(v, e)
	}
	if v, e := IvyInitAnd(IvyInit1, IvyInitDC); e != nil || v != IvyInitDC {
		t.Fatal(v, e)
	}
	if v, e := IvyInitXor(IvyInit1, IvyInit0); e != nil || v != IvyInit1 {
		t.Fatal(v, e)
	}
	leaf, e := IvyLeaf(123, 17)
	if e != nil || IvyLeafID(leaf) != 123 || IvyLeafLatency(leaf) != 17 {
		t.Fatal(leaf, e)
	}
}

func TestIvySequentialDFSCloneCleanupAndOwnership(t *testing.T) {
	m, _ := NewIvyManager(IvyOptions{})
	a := mustIvy(m.CreatePI())
	b := mustIvy(m.CreatePI())
	n := mustIvy(m.And(a, b))
	l := mustIvy(m.Latch(n, IvyInit0))
	q := mustIvy(m.And(l, a))
	if _, err := m.CreatePO(q); err != nil {
		t.Fatal(err)
	}
	dangling := mustIvy(m.And(a.Not(), b))
	nodes, latches, err := m.SequentialDFS()
	if err != nil {
		t.Fatal(err)
	}
	if len(latches) != 1 || latches[0] != l.ID() {
		t.Fatalf("latches %v", latches)
	}
	if len(nodes) != 2 {
		t.Fatalf("reachable nodes %v", nodes)
	}
	fan, err := m.Fanouts(a.ID())
	if err != nil {
		t.Fatal(err)
	}
	if len(fan) != 3 {
		t.Fatalf("fanouts %v", fan)
	}
	clone, err := m.Clone()
	if err != nil {
		t.Fatal(err)
	}
	removed, err := m.Cleanup()
	if err != nil || removed != 1 {
		t.Fatalf("cleanup %d %v", removed, err)
	}
	if _, err = m.Object(dangling.ID()); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("dangling survived: %v", err)
	}
	if _, err = clone.Object(dangling.ID()); err != nil {
		t.Fatalf("clone aliased: %v", err)
	}
	if err = m.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestIvyLevelsLimitsCancellationAndErrors(t *testing.T) {
	m, _ := NewIvyManager(IvyOptions{MaxObjects: 4})
	var err error
	a := mustIvy(m.CreatePI())
	b := mustIvy(m.CreatePI())
	n := mustIvy(m.And(a, b))
	if _, err := m.CreatePO(n); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit %v", err)
	}
	r, _ := NewIvyManager(IvyOptions{MaxObjects: 4})
	ra := mustIvy(r.CreatePI())
	rb := mustIvy(r.CreatePI())
	if _, err := r.Xor(ra, rb); !errors.Is(err, ErrLimit) || r.LiveCount() != 3 {
		t.Fatalf("transactional limit: live=%d err=%v", r.LiveCount(), err)
	}
	if _, err = m.And(IvyEdge(9999), a); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("literal %v", err)
	}
	if _, err = m.Multi(nil, IvyAND); !errors.Is(err, ErrDimension) {
		t.Fatalf("multi %v", err)
	}
	if _, err = m.Latch(a.Not(), IvyInit0); !errors.Is(err, ErrDimension) {
		t.Fatalf("latch complement %v", err)
	}
	cancelled := false
	c, _ := NewIvyManager(IvyOptions{Cancelled: func() bool { return cancelled }})
	x := mustIvy(c.CreatePI())
	cancelled = true
	if _, err = c.And(x, IvyConst1); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	l, _ := NewIvyManager(IvyOptions{})
	x = mustIvy(l.CreatePI())
	y := mustIvy(l.CreatePI())
	z := mustIvy(l.And(x, y))
	if _, err = l.CreatePO(z); err != nil {
		t.Fatal(err)
	}
	levels, err := l.Levelize()
	if err != nil || len(levels) < 2 || len(levels[1]) != 1 {
		t.Fatalf("levels %v %v", levels, err)
	}
	required, err := l.RequiredLevels()
	if err != nil || required[z.ID()] != 1 {
		t.Fatalf("required %v %v", required, err)
	}
}

func TestIvyConeAndDeepIterativeTraversal(t *testing.T) {
	m, _ := NewIvyManager(IvyOptions{MaxObjects: 5000})
	root := mustIvy(m.CreatePI())
	front := root
	for i := 0; i < 1500; i++ {
		v := mustIvy(m.CreatePI())
		root = mustIvy(m.And(root, v))
		if i == 749 {
			front = root
		}
	}
	if _, err := m.CreatePO(root); err != nil {
		t.Fatal(err)
	}
	nodes, err := m.DFS()
	if err != nil || len(nodes) != 1500 {
		t.Fatalf("dfs %d %v", len(nodes), err)
	}
	cone, err := m.CollectCone(root, []IvyEdge{front})
	if err != nil || len(cone) != 750 {
		t.Fatalf("cone %d %v", len(cone), err)
	}
	if err = m.Check(); err != nil {
		t.Fatal(err)
	}
}

func ivyModelEval(objects []IvyObject, outputs []uint32, inputs []uint32, assignment []bool) []bool {
	values := make([]bool, len(objects))
	values[0] = true
	for i, id := range inputs {
		values[id] = assignment[i]
	}
	read := func(e IvyEdge) bool {
		v := values[e.ID()]
		if e.Complemented() {
			return !v
		}
		return v
	}
	for id, o := range objects {
		if id == 0 || !o.Alive {
			continue
		}
		switch o.Type {
		case IvyAND:
			values[id] = read(o.Fanin0) && read(o.Fanin1)
		case IvyEXOR:
			values[id] = read(o.Fanin0) != read(o.Fanin1)
		case IvyBuffer, IvyPO:
			values[id] = read(o.Fanin0)
		}
	}
	r := make([]bool, len(outputs))
	for i, id := range outputs {
		r[i] = values[id]
	}
	return r
}

func FuzzIvyGraphModel(f *testing.F) {
	f.Add([]byte{2, 0, 1, 2, 3, 4, 5, 6})
	f.Add([]byte{0})
	f.Add([]byte{4, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		m, err := NewIvyManager(IvyOptions{MaxObjects: 400})
		if err != nil {
			t.Fatal(err)
		}
		signals := []IvyEdge{IvyConst0, IvyConst1}
		nInputs := 1
		if len(data) > 0 {
			nInputs = 1 + int(data[0]%5)
		}
		for range nInputs {
			signals = append(signals, mustIvy(m.CreatePI()))
		}
		for _, op := range data {
			a := signals[int(op)%len(signals)]
			b := signals[int(op/3+1)%len(signals)]
			var r IvyEdge
			switch op % 5 {
			case 0:
				r, err = m.And(a, b)
			case 1:
				r, err = m.Or(a, b)
			case 2:
				r, err = m.Xor(a, b)
			case 3:
				r, err = m.Mux(a, b, a.Not())
			default:
				r, err = m.Majority(a, b, IvyConst1)
			}
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, r)
		}
		for i := 0; i < 3 && i < len(signals); i++ {
			if _, err = m.CreatePO(signals[len(signals)-1-i]); err != nil {
				t.Fatal(err)
			}
		}
		if err = m.Check(); err != nil {
			t.Fatal(err)
		}
		objects := make([]IvyObject, len(m.objects))
		copy(objects, m.objects)
		for mask := 0; mask < 1<<nInputs; mask++ {
			assignment := make([]bool, nInputs)
			for i := range assignment {
				assignment[i] = mask>>i&1 != 0
			}
			got, e := m.Evaluate(assignment, nil)
			if e != nil {
				t.Fatal(e)
			}
			want := ivyModelEval(objects, m.outputs, m.inputs, assignment)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("mask %d got %v want %v", mask, got, want)
			}
		}
	})
}

func TestIvyDeterministicStateModel(t *testing.T) {
	r := rand.New(rand.NewSource(89))
	for run := 0; run < 5000; run++ {
		m, _ := NewIvyManager(IvyOptions{MaxObjects: 100})
		signals := []IvyEdge{IvyConst0, IvyConst1}
		for range 3 {
			signals = append(signals, mustIvy(m.CreatePI()))
		}
		for range 12 {
			a := signals[r.Intn(len(signals))]
			b := signals[r.Intn(len(signals))]
			var e IvyEdge
			var err error
			switch r.Intn(3) {
			case 0:
				e, err = m.And(a, b)
			case 1:
				e, err = m.Or(a, b)
			default:
				e, err = m.Xor(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, e)
		}
		if _, err := m.CreatePO(signals[r.Intn(len(signals))]); err != nil {
			t.Fatal(err)
		}
		if err := m.Check(); err != nil {
			t.Fatal(err)
		}
		assignment := []bool{r.Intn(2) != 0, r.Intn(2) != 0, r.Intn(2) != 0}
		got, err := m.Evaluate(assignment, nil)
		if err != nil {
			t.Fatal(err)
		}
		want := ivyModelEval(m.objects, m.outputs, m.inputs, assignment)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d", run)
		}
	}
}
