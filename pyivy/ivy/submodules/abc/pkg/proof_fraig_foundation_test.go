package abc

// Tests for the stand-alone FRAIG owned-data foundation.

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func mustFraigManager(t *testing.T, options FraigOptions) *FraigManager {
	t.Helper()
	m, err := NewFraigManager(options)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustFraigInput(t *testing.T, m *FraigManager) FraigEdge {
	t.Helper()
	e, err := m.NewInput()
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func mustFraigAnd(t *testing.T, m *FraigManager, a, b FraigEdge) FraigEdge {
	t.Helper()
	e, err := m.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestFraigFoundationCanonicalTruthAndOwnership(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{FunctionalReduce: true})
	defer m.Close()
	x := mustFraigInput(t, m)
	y := mustFraigInput(t, m)
	z := mustFraigInput(t, m)
	xy := mustFraigAnd(t, m, x, y)
	if got := mustFraigAnd(t, m, y, x); got != xy {
		t.Fatalf("commuted structural node %v != %v", got, xy)
	}
	if got := mustFraigAnd(t, m, xy, x); got != xy {
		t.Fatalf("functional absorption did not canonicalize: %v != %v", got, xy)
	}
	xyz := mustFraigAnd(t, m, xy, z)
	if err := m.AddOutput(xyz.Not()); err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, err := m.Evaluate(xyz.Not(), in)
		if err != nil {
			t.Fatal(err)
		}
		want := !(in[0] && in[1] && in[2])
		if got != want {
			t.Fatalf("assignment %d got %t want %t", assignment, got, want)
		}
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	s, err := m.Snapshot(xy.ID())
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != 1 || s.References < 1 || len(s.Fanouts) != 1 {
		t.Fatalf("unexpected snapshot %+v", s)
	}
	s.Fanouts[0] = 999
	s2, _ := m.Snapshot(xy.ID())
	if s2.Fanouts[0] == 999 {
		t.Fatal("snapshot aliases manager")
	}
	outs := m.Outputs()
	outs[0] = 0
	if m.Outputs()[0] != xyz.Not() {
		t.Fatal("outputs snapshot aliases manager")
	}
	if _, err = m.RemoveLastOutput(); err != nil {
		t.Fatal(err)
	}
	if err = m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFraigTruthExpansionAfterLateInput(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{FunctionalReduce: true})
	x := mustFraigInput(t, m)
	if v, e := m.Evaluate(x, []bool{true}); e != nil || !v {
		t.Fatalf("before expansion %t %v", v, e)
	}
	_ = mustFraigInput(t, m)
	for _, in := range [][]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		v, e := m.Evaluate(x, in)
		if e != nil || v != in[0] {
			t.Fatalf("expanded x on %v = %t, %v", in, v, e)
		}
	}
}

func TestFraigSimulationPrimitives(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	a := mustFraigInput(t, m)
	b := mustFraigInput(t, m)
	if err := m.SetSimulation(a.ID(), []uint32{0x0f, 0xaa}, []uint32{3}, false); err != nil {
		t.Fatal(err)
	}
	if err := m.SetSimulation(b.ID(), []uint32{0xf0, 0xaa}, []uint32{3}, false); err != nil {
		t.Fatal(err)
	}
	equal, err := m.CompareSimulation(a.ID(), b.ID(), true, 2)
	if err != nil || equal {
		t.Fatalf("comparison %t %v", equal, err)
	}
	diff, err := m.FirstSimulationDifference(a.ID(), b.ID(), false, true, 2)
	if err != nil || diff != 0 {
		t.Fatalf("first diff %d %v", diff, err)
	}
	x, err := m.SimulationXOR(a.ID(), b.ID(), true, 2)
	if err != nil || !reflect.DeepEqual(x, []uint32{0xff, 0}) {
		t.Fatalf("xor %x %v", x, err)
	}
	equal, err = m.CompareSimulation(a.ID(), b.ID(), false, 1)
	if err != nil || !equal {
		t.Fatalf("dynamic comparison %t %v", equal, err)
	}
	if err = m.SetSimulation(a.ID(), []uint32{0x55}, nil, false); err != nil {
		t.Fatal(err)
	}
	if err = m.SetSimulation(b.ID(), []uint32{0x55}, nil, true); err != nil {
		t.Fatal(err)
	}
	equal, err = m.CompareSimulation(a.ID(), b.ID(), true, 1)
	if err != nil || !equal {
		t.Fatalf("raw phase-normalized comparison %t %v", equal, err)
	}
	diff, err = m.FirstSimulationDifference(a.ID(), b.ID(), false, true, 1)
	if err != nil || diff != 0 {
		t.Fatalf("phase-aware first difference %d %v", diff, err)
	}
	equal, err = m.CompareSimulationUnderMask(a.ID(), b.ID(), true, []uint32{0xff})
	if err != nil || !equal {
		t.Fatalf("masked equality %t %v", equal, err)
	}
	if err = m.SetSimulation(b.ID(), []uint32{0x54}, nil, true); err != nil {
		t.Fatal(err)
	}
	equal, err = m.CompareSimulationUnderMask(a.ID(), b.ID(), true, []uint32{0xfe})
	if err != nil || !equal {
		t.Fatalf("masked ignored difference %t %v", equal, err)
	}
	equal, err = m.CompareSimulationUnderMask(a.ID(), b.ID(), true, []uint32{1})
	if err != nil || equal {
		t.Fatalf("masked selected difference %t %v", equal, err)
	}
}

func TestFraigDisabledFunctionalTableStartsEmpty(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	if got := m.Stats().FunctionalEntries; got != 0 {
		t.Fatalf("disabled functional table contains %d entries", got)
	}
}

func TestFraigFanoutTransfer(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	x := mustFraigInput(t, m)
	y := mustFraigInput(t, m)
	z := mustFraigInput(t, m)
	yz := mustFraigAnd(t, m, y, z)
	from := mustFraigAnd(t, m, x, yz)
	xy := mustFraigAnd(t, m, x, y)
	to := mustFraigAnd(t, m, xy, z)
	consumer := mustFraigAnd(t, m, from, y)
	if err := m.TransferFanouts(from.ID(), to.ID()); err != nil {
		t.Fatal(err)
	}
	s, err := m.Snapshot(consumer.ID())
	if err != nil {
		t.Fatal(err)
	}
	if s.Fanins[0].ID() != to.ID() && s.Fanins[1].ID() != to.ID() {
		t.Fatalf("fanout was not transferred: %+v", s)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, err := m.Evaluate(consumer, in)
		if err != nil {
			t.Fatal(err)
		}
		if got != (in[0] && in[1] && in[2]) {
			t.Fatalf("assignment %d changed after transfer", assignment)
		}
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFraigFanoutTransferCancellationIsAtomic(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	x := mustFraigInput(t, m)
	y := mustFraigInput(t, m)
	z := mustFraigInput(t, m)
	yz := mustFraigAnd(t, m, y, z)
	from := mustFraigAnd(t, m, x, yz)
	xy := mustFraigAnd(t, m, x, y)
	to := mustFraigAnd(t, m, xy, z)
	_ = mustFraigAnd(t, m, from, y)
	before := make([]FraigNodeSnapshot, m.Stats().Nodes)
	for id := range before {
		before[id], _ = m.Snapshot(FraigID(id))
	}
	calls := 0
	m.opts.Stop = func() bool {
		calls++
		// Initial check, one transferred fanout, and every node rebuild check
		// succeed. Cancellation is first observed by post-commit validation,
		// which must restore the original manager.
		return calls == len(before)+3
	}
	if err := m.TransferFanouts(from.ID(), to.ID()); !errors.Is(err, ErrCancelled) {
		t.Fatalf("transfer cancellation %v", err)
	}
	m.opts.Stop = nil
	for id, want := range before {
		got, err := m.Snapshot(FraigID(id))
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("node %d changed after cancellation: got %+v want %+v err %v", id, got, want, err)
		}
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFraigNodeVector(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	a := mustFraigInput(t, m)
	b := mustFraigInput(t, m)
	n := mustFraigAnd(t, m, a, b)
	v, err := NewFraigNodeVector(1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err = v.Push(n); err != nil {
		t.Fatal(err)
	}
	if old, e := v.PushUnique(n); e != nil || !old {
		t.Fatalf("unique %t %v", old, e)
	}
	_ = v.Push(a)
	_ = v.Push(b)
	if err = v.SortByLevel(m, true); err != nil {
		t.Fatal(err)
	}
	got := v.Snapshot()
	if got[2] != n {
		t.Fatalf("level order %v", got)
	}
	v.SortByNumber()
	got = v.Snapshot()
	if got[0].ID() > got[1].ID() {
		t.Fatalf("number order %v", got)
	}
	if err = v.SortByReferences(m); err != nil {
		t.Fatal(err)
	}
	copyOut := v.Snapshot()
	copyOut[0] = 999
	if v.Snapshot()[0] == 999 {
		t.Fatal("snapshot aliases vector")
	}
	if err = v.Push(a); err != nil {
		t.Fatal(err)
	}
	if err = v.Push(a); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit error %v", err)
	}
	if err = v.Remove(999); !errors.Is(err, ErrDimension) {
		t.Fatalf("remove error %v", err)
	}
	v.Close()
	if err = v.Push(a); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed error %v", err)
	}
}

func TestFraigNodeVectorOrderedInsertion(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{})
	a := mustFraigInput(t, m)
	b := mustFraigInput(t, m)
	n := mustFraigAnd(t, m, a, b)
	byNumber, _ := NewFraigNodeVector(0, 0)
	for _, e := range []FraigEdge{n, b, a} {
		if err := byNumber.PushOrderedNumber(e); err != nil {
			t.Fatal(err)
		}
	}
	if got := byNumber.Snapshot(); got[0] != a || got[1] != b || got[2] != n {
		t.Fatalf("number insertion order %v", got)
	}
	if old, err := byNumber.PushUniqueOrderedNumber(b); err != nil || !old {
		t.Fatalf("ordered uniqueness %t %v", old, err)
	}
	clone, err := byNumber.Clone()
	if err != nil {
		t.Fatal(err)
	}
	_ = clone.Write(0, n)
	if byNumber.Snapshot()[0] == n {
		t.Fatal("clone aliases source")
	}
	byLevel, _ := NewFraigNodeVector(0, 0)
	for _, e := range []FraigEdge{a, n, b} {
		if err := byLevel.PushOrderedLevel(m, e); err != nil {
			t.Fatal(err)
		}
	}
	if got := byLevel.Snapshot(); got[0] != n {
		t.Fatalf("level insertion order %v", got)
	}
}

func TestFraigFixedPoolLifecycle(t *testing.T) {
	p, err := NewFraigFixedPool(16, 10)
	if err != nil {
		t.Fatal(err)
	}
	handles := make([]int, 10)
	for i := range handles {
		handles[i], err = p.Fetch()
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = p.Fetch(); !errors.Is(err, ErrLimit) {
		t.Fatalf("pool limit %v", err)
	}
	if err = p.Recycle(handles[4]); err != nil {
		t.Fatal(err)
	}
	h, err := p.Fetch()
	if err != nil || h != handles[4] {
		t.Fatalf("recycle returned %d %v", h, err)
	}
	if err = p.Recycle(99); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad recycle %v", err)
	}
	stats := p.Stats()
	if stats.Used != 10 || stats.Peak != 10 || stats.Allocated != 10 || stats.Memory != 160 {
		t.Fatalf("stats %+v", stats)
	}
	if err = p.Restart(); err != nil {
		t.Fatal(err)
	}
	if p.Stats().Used != 0 {
		t.Fatal("restart retained used entries")
	}
	p.Close()
	if _, err = p.Fetch(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed pool %v", err)
	}
}

func TestFraigFixedPoolRejectsOverflowingEntrySize(t *testing.T) {
	if _, err := NewFraigFixedPool(int(^uint(0)>>1), 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("overflowing entry size error %v", err)
	}
}

func TestFraigLimitsCancellationAndAtomicity(t *testing.T) {
	m := mustFraigManager(t, FraigOptions{MaxNodes: 2, MaxInputs: 2})
	a := mustFraigInput(t, m)
	before := m.Stats()
	if _, err := m.NewInput(); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit %v", err)
	}
	if m.Stats() != before {
		t.Fatal("failed input mutated manager")
	}
	if _, err := m.And(a, a.Not()); err != nil {
		t.Fatal(err)
	}
	calls := 0
	c := mustFraigManager(t, FraigOptions{Stop: func() bool { calls++; return calls > 1 }})
	if _, err := c.NewInput(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation %v", err)
	}
	if c.Stats().Nodes != 1 {
		t.Fatal("cancelled input mutated manager")
	}
	c.Close()
	if _, err := c.NewInput(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager %v", err)
	}
}

func TestFraigIndependentStateModel5000(t *testing.T) {
	r := rand.New(rand.NewSource(0x4652414947))
	m := mustFraigManager(t, FraigOptions{MaxTruthInputs: 6})
	edges := []FraigEdge{m.ConstantOne(), m.ConstantZero()}
	truths := []uint64{^uint64(0), 0}
	inputs := 0
	for step := 0; step < 5000; step++ {
		if inputs < 6 && (len(edges) < 3 || r.Intn(5) == 0) {
			e, err := m.NewInput()
			if err != nil {
				t.Fatal(err)
			}
			oldWidth := 1 << inputs
			mask := fraigTruthMask(inputs + 1)
			for i := range truths {
				truths[i] = (truths[i] | truths[i]<<oldWidth) & mask
			}
			var tv uint64
			for a := 0; a < 1<<(inputs+1); a++ {
				if a&(1<<inputs) != 0 {
					tv |= uint64(1) << a
				}
			}
			edges = append(edges, e)
			truths = append(truths, tv)
			inputs++
			continue
		}
		i, j := r.Intn(len(edges)), r.Intn(len(edges))
		a, b := edges[i], edges[j]
		ta, tb := truths[i], truths[j]
		if r.Intn(2) == 0 {
			a = a.Not()
			ta = (^ta) & fraigTruthMask(inputs)
		}
		if r.Intn(2) == 0 {
			b = b.Not()
			tb = (^tb) & fraigTruthMask(inputs)
		}
		e, err := m.And(a, b)
		if err != nil {
			t.Fatal(err)
		}
		want := ta & tb
		for assignment := 0; assignment < 1<<inputs; assignment++ {
			in := make([]bool, inputs)
			for k := range in {
				in[k] = assignment&(1<<k) != 0
			}
			got, err := m.Evaluate(e, in)
			if err != nil {
				t.Fatal(err)
			}
			if got != (want&(uint64(1)<<assignment) != 0) {
				t.Fatalf("step %d assignment %d", step, assignment)
			}
		}
		edges = append(edges, e)
		truths = append(truths, want)
		if len(edges) > 128 {
			edges = edges[len(edges)-96:]
			truths = truths[len(truths)-96:]
		}
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func FuzzFraigFoundation(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{0xff, 0, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		m, err := NewFraigManager(FraigOptions{MaxNodes: 300, MaxInputs: 6, FunctionalReduce: true})
		if err != nil {
			t.Fatal(err)
		}
		edges := []FraigEdge{m.ConstantOne(), m.ConstantZero()}
		truths := []uint64{1, 0}
		inputs := 0
		for _, op := range data {
			if op&7 == 0 && inputs < 6 {
				e, err := m.NewInput()
				if err != nil {
					t.Fatal(err)
				}
				oldWidth := 1 << inputs
				mask := fraigTruthMask(inputs + 1)
				for i := range truths {
					truths[i] = (truths[i] | truths[i]<<oldWidth) & mask
				}
				var tv uint64
				for a := 0; a < 1<<(inputs+1); a++ {
					if a&(1<<inputs) != 0 {
						tv |= uint64(1) << a
					}
				}
				edges = append(edges, e)
				truths = append(truths, tv)
				inputs++
				continue
			}
			if len(edges) == 0 {
				continue
			}
			i := int(op) % len(edges)
			j := int(op>>3) % len(edges)
			a, b := edges[i], edges[j]
			ta, tb := truths[i], truths[j]
			mask := fraigTruthMask(inputs)
			if op&0x40 != 0 {
				a = a.Not()
				ta = (^ta) & mask
			}
			if op&0x80 != 0 {
				b = b.Not()
				tb = (^tb) & mask
			}
			e, err := m.And(a, b)
			if errors.Is(err, ErrLimit) {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			want := ta & tb
			edges = append(edges, e)
			truths = append(truths, want)
		}
		for i, e := range edges {
			if inputs > 6 {
				break
			}
			for assignment := 0; assignment < 1<<inputs; assignment++ {
				in := make([]bool, inputs)
				for k := range in {
					in[k] = assignment&(1<<k) != 0
				}
				got, err := m.Evaluate(e, in)
				if err != nil {
					t.Fatal(err)
				}
				if got != (truths[i]&(uint64(1)<<assignment) != 0) {
					t.Fatalf("edge %d assignment %d", i, assignment)
				}
			}
		}
		if err := m.Validate(); err != nil {
			t.Fatal(err)
		}
	})
}
