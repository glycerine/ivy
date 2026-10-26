package abc

// Tests for dynamically maintained AIG order from aig/aig/aigOrder.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigOrderFixtureState struct {
	Previous, Next AigID
	Before, Total  int
	Sequence       []AigID
}

func aigOrderFixture(t testing.TB, seed int) (*AigManager, []AigLit) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed>>uint(i%5)&1 != 0)
	}
	node := make([]AigLit, 6)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[2], leaf[3])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(leaf[4], leaf[5])
	node[4], _ = g.And(node[2], node[3])
	node[5], _ = g.And(node[0], node[3])
	return g, node
}

func aigOrderSnapshot(t testing.TB, order *AigOrder) aigOrderFixtureState {
	t.Helper()
	sequence, err := order.Sequence()
	if err != nil {
		t.Fatal(err)
	}
	state := order.State()
	return aigOrderFixtureState{state.Previous, state.Next, state.Before, state.Total, sequence}
}

func TestAigOrderNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_order.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Steps                          int
		InsertedID, RemovedPrevID            AigID
		RemovedNextID                        AigID
		Initial, Begun, Inserted, Selected   aigOrderFixtureState
		RemovedNew, RemovedPrev, RemovedNext aigOrderFixtureState
		Final                                aigOrderFixtureState
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, nodes := aigOrderFixture(t, row.Seed)
		order, buildErr := BuildAigOrder(g, AigTraversalOptions{})
		if buildErr != nil || !reflect.DeepEqual(aigOrderSnapshot(t, order), row.Initial) {
			t.Fatal(row.Seed, buildErr, aigOrderSnapshot(t, order), row.Initial)
		}
		if err = order.Begin(); err != nil || !reflect.DeepEqual(aigOrderSnapshot(t, order), row.Begun) {
			t.Fatal(row.Seed, err, aigOrderSnapshot(t, order), row.Begun)
		}
		for i := 0; i < row.Steps; i++ {
			if err = order.Step(); err != nil {
				t.Fatal(row.Seed, i, err)
			}
		}
		inserted, _ := g.And(nodes[4], nodes[5])
		if inserted.ID() != row.InsertedID || order.Insert(g, inserted.ID()) != nil || !reflect.DeepEqual(aigOrderSnapshot(t, order), row.Inserted) {
			t.Fatal(row.Seed, inserted, aigOrderSnapshot(t, order), row.Inserted)
		}
		if err = order.Refresh(); err != nil || !reflect.DeepEqual(aigOrderSnapshot(t, order), row.Selected) {
			t.Fatal(row.Seed, err, aigOrderSnapshot(t, order), row.Selected)
		}
		if err = order.Advance(); err != nil {
			t.Fatal(row.Seed, err)
		}
		if err = order.Remove(g, inserted.ID()); err != nil || !reflect.DeepEqual(aigOrderSnapshot(t, order), row.RemovedNew) {
			t.Fatal(row.Seed, err, aigOrderSnapshot(t, order), row.RemovedNew)
		}
		if row.RemovedPrevID != 0 {
			if err = order.Remove(g, row.RemovedPrevID); err != nil {
				t.Fatal(row.Seed, err)
			}
		}
		if !reflect.DeepEqual(aigOrderSnapshot(t, order), row.RemovedPrev) {
			t.Fatal(row.Seed, aigOrderSnapshot(t, order), row.RemovedPrev)
		}
		if row.RemovedNextID != 0 {
			if err = order.Remove(g, row.RemovedNextID); err != nil {
				t.Fatal(row.Seed, err)
			}
		}
		if !reflect.DeepEqual(aigOrderSnapshot(t, order), row.RemovedNext) {
			t.Fatal(row.Seed, aigOrderSnapshot(t, order), row.RemovedNext)
		}
		if err = order.Refresh(); err != nil {
			t.Fatal(row.Seed, err)
		}
		for {
			if _, ok := order.Current(); !ok {
				break
			}
			if err = order.Step(); err != nil {
				t.Fatal(row.Seed, err)
			}
		}
		if !reflect.DeepEqual(aigOrderSnapshot(t, order), row.Final) {
			t.Fatal(row.Seed, aigOrderSnapshot(t, order), row.Final)
		}
	}
}

func TestAigOrderOwnershipErrorsLimitsAndWide(t *testing.T) {
	g, nodes := aigOrderFixture(t, 0)
	order, err := BuildAigOrder(g, AigTraversalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	clone := order.Clone()
	if clone.Begin() != nil || clone.Step() != nil || clone.Remove(g, nodes[0].ID()) != nil || reflect.DeepEqual(clone, order) {
		t.Fatal(clone, order)
	}
	before := order.Clone()
	for _, id := range []AigID{0, g.cis[0], AigID(g.ObjectCount()), nodes[0].ID()} {
		if err = order.Insert(g, id); err == nil || !reflect.DeepEqual(order, before) {
			t.Fatal(id, err)
		}
	}
	newNode, _ := g.And(nodes[4], nodes[5])
	if err = order.Remove(g, newNode.ID()); !errors.Is(err, ErrInvariant) || !reflect.DeepEqual(order, before) {
		t.Fatal(err)
	}
	if sequence, err := (*AigOrder)(nil).Sequence(); sequence != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(sequence, err)
	}
	for _, options := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if result, err := BuildAigOrder(g, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	if result, err := BuildAigOrder(buffered, AigTraversalOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	empty := NewAigManager()
	emptyOrder, err := BuildAigOrder(empty, AigTraversalOptions{})
	if err != nil || emptyOrder.Begin() != nil {
		t.Fatal(err)
	}
	if _, ok := emptyOrder.Current(); ok {
		t.Fatal("empty current")
	}
	if err = emptyOrder.Advance(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	const width = 10000
	wide := NewAigManager()
	common, _ := wide.CreateCI()
	for i := 0; i < width; i++ {
		other, _ := wide.CreateCI()
		wide.And(common, other)
	}
	wideOrder, err := BuildAigOrder(wide, AigTraversalOptions{MaxObjects: 30000, MaxVisits: 30000})
	if err != nil {
		t.Fatal(err)
	}
	sequence, err := wideOrder.Sequence()
	if err != nil || len(sequence) != width {
		t.Fatal(len(sequence), err)
	}
	for i, id := range sequence {
		if id != AigID(3+2*i) {
			t.Fatal(i, id)
		}
	}
}

func FuzzAigOrder(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 6)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		for i := 0; i+1 < len(data) && len(signals) < 70; i += 2 {
			a := signals[int(data[i]&0x7f)%len(signals)].NotIf(data[i]&0x80 != 0)
			b := signals[int(data[i+1]&0x7f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			if value, err := g.And(a, b); err == nil {
				signals = append(signals, value)
			}
		}
		order, err := BuildAigOrder(g, AigTraversalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := make([]AigID, 0)
		for id, object := range g.objects {
			if object.Type == AigAND {
				want = append(want, AigID(id))
			}
		}
		sequence, err := order.Sequence()
		if err != nil || !slices.Equal(sequence, want) {
			t.Fatal(sequence, want, err)
		}
		if err = order.Begin(); err != nil {
			t.Fatal(err)
		}
		step := 0
		for {
			current, ok := order.Current()
			if !ok {
				break
			}
			code := byte(step)
			if len(data) > 0 {
				code = data[step%len(data)]
			}
			if code%3 == 0 {
				if err = order.Remove(g, current); err != nil {
					t.Fatal(err)
				}
				want = slices.DeleteFunc(want, func(id AigID) bool { return id == current })
			} else {
				if err = order.Step(); err != nil {
					t.Fatal(err)
				}
				if code%3 == 1 {
					if err = order.Remove(g, current); err != nil {
						t.Fatal(err)
					}
					want = slices.DeleteFunc(want, func(id AigID) bool { return id == current })
				}
			}
			step++
		}
		sequence, err = order.Sequence()
		state := order.State()
		if err != nil || !slices.Equal(sequence, want) || state.Next != 0 || state.Before != state.Total || state.Total != len(want) {
			t.Fatal(sequence, want, state, err)
		}
	})
}
