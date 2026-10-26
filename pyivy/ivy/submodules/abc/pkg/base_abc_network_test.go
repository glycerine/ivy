package abc

// Tests for the Go port of src/base/abc/abc.h, abcInt.h, abcObj.c, abcNtk.c and abcFanio.c.

import (
	"errors"
	"math/rand"
	"slices"
	"sync"
	"testing"
)

func newTestABCNetwork(t *testing.T) *ABCNetwork {
	t.Helper()
	n, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestABCNetworkPartitionsAndNames(t *testing.T) {
	n := newTestABCNetwork(t)
	pi, _ := n.CreatePI()
	po, _ := n.CreatePO()
	bi, latch, bo, err := n.CreateLatchBox(ABCInitZero)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(n.PIs(), []int{pi}) || !slices.Equal(n.POs(), []int{po}) || !slices.Equal(n.CIs(), []int{pi, bo}) || !slices.Equal(n.COs(), []int{po, bi}) || !slices.Equal(n.Boxes(), []int{latch}) {
		t.Fatalf("bad partitions: PI=%v PO=%v CI=%v CO=%v boxes=%v", n.PIs(), n.POs(), n.CIs(), n.COs(), n.Boxes())
	}
	if err = n.SetObjectName(pi, "a"); err != nil {
		t.Fatal(err)
	}
	if got, ok := n.FindObject("a"); !ok || got != pi {
		t.Fatalf("name lookup %d %v", got, ok)
	}
	if err = n.SetObjectName(po, "a"); !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate name: %v", err)
	}
	if err = n.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestABCNetworkFanioAndOwnership(t *testing.T) {
	n := newTestABCNetwork(t)
	a, _ := n.CreatePI()
	b, _ := n.CreatePI()
	node, err := n.CreateAND(ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b, Complemented: true})
	if err != nil {
		t.Fatal(err)
	}
	po, _ := n.CreatePO()
	if err = n.AddFanin(po, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	if err = n.AddFanin(po, ABCObjectEdge{ID: a}); !errors.Is(err, ErrDimension) {
		t.Fatalf("second CO fanin: %v", err)
	}
	inserted, err := n.InsertBetween(node, po, ABCObjectNode)
	if err != nil {
		t.Fatal(err)
	}
	if err = n.PatchFanin(inserted, node, ABCObjectEdge{ID: a, Complemented: true}); err != nil {
		t.Fatal(err)
	}
	if err = n.TransferFanouts(inserted, b); err != nil {
		t.Fatal(err)
	}
	if err = n.DeleteObject(inserted); err != nil {
		t.Fatal(err)
	}
	if err = n.Validate(); err != nil {
		t.Fatal(err)
	}
	o, _ := n.Object(node)
	o.Fanins[0].ID = 999
	o2, _ := n.Object(node)
	if o2.Fanins[0].ID == 999 {
		t.Fatal("snapshot aliases manager")
	}
	c := n.Clone()
	if err = c.SetObjectName(node, "clone"); err != nil {
		t.Fatal(err)
	}
	if _, ok := n.FindObject("clone"); ok {
		t.Fatal("clone name aliases source")
	}
	idx, err := n.FanoutFaninIndex(po, b)
	if err != nil || idx != 0 {
		t.Fatalf("fanin index %d %v", idx, err)
	}
	constant, err := n.ReplaceObjectByConstant(node, true)
	if err != nil {
		t.Fatal(err)
	}
	co, _ := n.Object(constant)
	if co.Function != " 1\n" {
		t.Fatalf("constant function %q", co.Function)
	}
}

func TestABCNetworkPatchFaninPhaseAndInsert(t *testing.T) {
	n := newTestABCNetwork(t)
	a, _ := n.CreatePI()
	b, _ := n.CreatePI()
	node, err := n.CreateAND(ABCObjectEdge{ID: a, Complemented: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = n.PatchFanin(node, a, ABCObjectEdge{ID: b}); err != nil {
		t.Fatal(err)
	}
	got, _ := n.Object(node)
	if len(got.Fanins) != 1 || got.Fanins[0] != (ABCObjectEdge{ID: b, Complemented: true}) {
		t.Fatalf("patch did not preserve old phase: %+v", got.Fanins)
	}
	if err = n.PatchFanin(node, b, ABCObjectEdge{ID: a, Complemented: true}); err != nil {
		t.Fatal(err)
	}
	got, _ = n.Object(node)
	if got.Fanins[0] != (ABCObjectEdge{ID: a}) {
		t.Fatalf("patch did not XOR replacement phase: %+v", got.Fanins[0])
	}
	po, _ := n.CreatePO()
	if err = n.AddFanin(po, ABCObjectEdge{ID: node, Complemented: true}); err != nil {
		t.Fatal(err)
	}
	inserted, err := n.InsertBetween(node, po, ABCObjectNode)
	if err != nil {
		t.Fatal(err)
	}
	poObj, _ := n.Object(po)
	insertedObj, _ := n.Object(inserted)
	if poObj.Fanins[0] != (ABCObjectEdge{ID: inserted, Complemented: true}) || insertedObj.Fanins[0] != (ABCObjectEdge{ID: node}) {
		t.Fatalf("insert changed edge phases: po=%+v inserted=%+v", poObj.Fanins, insertedObj.Fanins)
	}
}

func TestABCNetworkStrashConstant(t *testing.T) {
	n, err := NewABCNetwork(ABCNetworkStrash, ABCFunctionAIG, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	o, err := n.Object(0)
	if err != nil || o.Type != ABCObjectConst1 || n.ObjectCountByType(ABCObjectConst1) != 1 {
		t.Fatalf("distinguished constant: %+v %v", o, err)
	}
	pi, err := n.CreatePI()
	if err != nil || pi != 1 {
		t.Fatalf("first strash PI = %d, %v", pi, err)
	}
	if err = n.AddFanin(0, ABCObjectEdge{ID: pi}); !errors.Is(err, ErrDimension) {
		t.Fatalf("constant accepted fanin: %v", err)
	}
	if err = n.DeleteObject(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("constant deletion: %v", err)
	}
	if err = n.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestABCNetworkCancellationIsAtomic(t *testing.T) {
	n := newTestABCNetwork(t)
	a, _ := n.CreatePI()
	b, _ := n.CreatePI()
	node, _ := n.CreateAND(ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b})
	calls := 0
	n.limits.Stop = func() bool {
		calls++
		return calls == 4
	}
	if err := n.RemoveFanins(node); !errors.Is(err, ErrCancelled) {
		t.Fatalf("remove cancellation: %v", err)
	}
	n.limits.Stop = nil
	o, _ := n.Object(node)
	if len(o.Fanins) != 2 {
		t.Fatalf("cancelled removal committed: %+v", o.Fanins)
	}

	old, _ := n.CreateNode()
	replacement, _ := n.CreateNode()
	po, _ := n.CreatePO()
	if err := n.AddFanin(po, ABCObjectEdge{ID: old, Complemented: true}); err != nil {
		t.Fatal(err)
	}
	n.limits.Stop = func() bool { return len(n.objects[old].Fanouts) == 0 }
	if err := n.ReplaceObject(old, replacement); !errors.Is(err, ErrCancelled) {
		t.Fatalf("replace cancellation: %v", err)
	}
	n.limits.Stop = nil
	oldObj, _ := n.Object(old)
	replacementObj, _ := n.Object(replacement)
	poObj, _ := n.Object(po)
	if !slices.Equal(oldObj.Fanouts, []int{po}) || len(replacementObj.Fanouts) != 0 || poObj.Fanins[0] != (ABCObjectEdge{ID: old, Complemented: true}) {
		t.Fatalf("cancelled replacement committed: old=%v new=%v po=%v", oldObj.Fanouts, replacementObj.Fanouts, poObj.Fanins)
	}
}

func TestABCNetworkAtomicLimitsCancellationLifecycle(t *testing.T) {
	n, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 3, MaxEdges: 1, MaxNameBytes: 2})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.CreatePI()
	b, _ := n.CreatePI()
	if _, err = n.CreateAND(ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b}); !errors.Is(err, ErrLimit) {
		t.Fatalf("edge limit: %v", err)
	}
	if n.ObjectCount() != 2 {
		t.Fatalf("failed create mutated manager: %d", n.ObjectCount())
	}
	if err = n.SetObjectName(a, "abc"); !errors.Is(err, ErrLimit) {
		t.Fatalf("name limit: %v", err)
	}
	n.Close()
	n.Close()
	if _, err = n.CreatePI(); !errors.Is(err, ErrABCNetworkClosed) {
		t.Fatalf("closed create: %v", err)
	}
	cancelled, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{Stop: func() bool { return true }})
	if _, err = cancelled.CreatePI(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
}

func TestABCNetworkObjectRules(t *testing.T) {
	n := newTestABCNetwork(t)
	pi, _ := n.CreatePI()
	po, _ := n.CreatePO()
	node, _ := n.CreateNode()
	net, _ := n.CreateNet()
	bo, _ := n.CreateBO()
	if err := n.AddFanin(pi, ABCObjectEdge{ID: node}); !errors.Is(err, ErrDimension) {
		t.Fatalf("PI accepted fanin: %v", err)
	}
	if err := n.AddFanin(node, ABCObjectEdge{ID: po}); !errors.Is(err, ErrDimension) {
		t.Fatalf("PO accepted fanout: %v", err)
	}
	if err := n.AddFanin(net, ABCObjectEdge{ID: pi}); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(net, ABCObjectEdge{ID: node}); !errors.Is(err, ErrDimension) {
		t.Fatalf("net accepted second driver: %v", err)
	}
	if err := n.AddFanin(bo, ABCObjectEdge{ID: node}); err != nil {
		t.Fatalf("BO should accept box driver: %v", err)
	}
	if err := n.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestABCNetworkIndependentManagers(t *testing.T) {
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
			if err != nil {
				t.Error(err)
				return
			}
			for i := 0; i < 100; i++ {
				if _, err = n.CreatePI(); err != nil {
					t.Error(err)
					return
				}
			}
			if err = n.Validate(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

type abcModelObject struct {
	alive   bool
	typ     ABCObjectType
	fanins  []int
	fanouts []int
}

func TestABCNetworkIndependentMutationModel(t *testing.T) {
	r := rand.New(rand.NewSource(0xabc5))
	for trial := 0; trial < 5000; trial++ {
		n := newTestABCNetwork(t)
		model := []abcModelObject{{}}
		for step := 0; step < 30; step++ {
			switch r.Intn(4) {
			case 0:
				types := []ABCObjectType{ABCObjectPI, ABCObjectPO, ABCObjectNode, ABCObjectNet}
				typ := types[r.Intn(len(types))]
				id, err := n.CreateObject(typ)
				if err != nil {
					t.Fatal(err)
				}
				for len(model) <= id {
					model = append(model, abcModelObject{})
				}
				model[id] = abcModelObject{alive: true, typ: typ}
			case 1:
				if len(model) < 3 {
					continue
				}
				dst := r.Intn(len(model))
				src := r.Intn(len(model))
				if dst == src || dst >= len(model) || src >= len(model) || !model[dst].alive || !model[src].alive {
					continue
				}
				if model[dst].typ == ABCObjectPI || model[src].typ == ABCObjectPO || (model[dst].typ == ABCObjectPO || model[dst].typ == ABCObjectBI || model[dst].typ == ABCObjectNet) && len(model[dst].fanins) > 0 {
					continue
				}
				if err := n.AddFanin(dst, ABCObjectEdge{ID: src, Complemented: r.Intn(2) == 1}); err != nil {
					t.Fatal(err)
				}
				model[dst].fanins = append(model[dst].fanins, src)
				model[src].fanouts = append(model[src].fanouts, dst)
			case 2:
				id := r.Intn(len(model))
				if id >= len(model) || !model[id].alive || len(model[id].fanins) == 0 {
					continue
				}
				src := model[id].fanins[0]
				if err := n.DeleteFanin(id, src); err != nil {
					t.Fatal(err)
				}
				model[id].fanins = model[id].fanins[1:]
				model[src].fanouts = removeModelOnce(model[src].fanouts, id)
			case 3:
				id := r.Intn(len(model))
				if id >= len(model) || !model[id].alive {
					continue
				}
				if err := n.DeleteObject(id); err != nil {
					t.Fatal(err)
				}
				for _, src := range model[id].fanins {
					model[src].fanouts = removeModelOnce(model[src].fanouts, id)
				}
				for _, dst := range model[id].fanouts {
					model[dst].fanins = removeModelOnce(model[dst].fanins, id)
				}
				model[id] = abcModelObject{}
			}
			if err := n.Validate(); err != nil {
				t.Fatalf("trial %d step %d: %v", trial, step, err)
			}
			for id, m := range model {
				if !m.alive {
					continue
				}
				got, err := n.Object(id)
				if err != nil {
					t.Fatal(err)
				}
				gf := make([]int, len(got.Fanins))
				for i, e := range got.Fanins {
					gf[i] = e.ID
				}
				if !slices.Equal(gf, m.fanins) || !slices.Equal(got.Fanouts, m.fanouts) {
					t.Fatalf("model mismatch trial=%d step=%d id=%d got=%v/%v want=%v/%v", trial, step, id, gf, got.Fanouts, m.fanins, m.fanouts)
				}
			}
		}
	}
}

func removeModelOnce(v []int, x int) []int {
	for i, y := range v {
		if y == x {
			return append(v[:i], v[i+1:]...)
		}
	}
	return v
}

func FuzzABCNetworkModel(f *testing.F) {
	f.Add([]byte{0, 2, 0, 2, 1, 1, 2, 0, 3, 2, 1})
	f.Add([]byte{0, 7, 0, 3, 1, 2, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		n := newTestABCNetwork(t)
		ids := make([]int, 0, 32)
		for i := 0; i < len(data); {
			op := data[i] % 5
			i++
			switch op {
			case 0:
				types := []ABCObjectType{ABCObjectPI, ABCObjectPO, ABCObjectNode, ABCObjectNet}
				typ := types[int(data[(i-1)%len(data)])%len(types)]
				id, err := n.CreateObject(typ)
				if err == nil {
					ids = append(ids, id)
				}
			case 1:
				if len(ids) < 2 || i+1 >= len(data) {
					continue
				}
				dst := ids[int(data[i])%len(ids)]
				src := ids[int(data[i+1])%len(ids)]
				i += 2
				_ = n.AddFanin(dst, ABCObjectEdge{ID: src, Complemented: data[i-1]&1 != 0})
			case 2:
				if len(ids) < 2 || i+1 >= len(data) {
					continue
				}
				dst := ids[int(data[i])%len(ids)]
				src := ids[int(data[i+1])%len(ids)]
				i += 2
				_ = n.DeleteFanin(dst, src)
			case 3:
				if len(ids) == 0 || i >= len(data) {
					continue
				}
				id := ids[int(data[i])%len(ids)]
				i++
				if n.DeleteObject(id) == nil {
					ids = slices.DeleteFunc(ids, func(x int) bool { return x == id })
				}
			case 4:
				if len(ids) < 3 || i+2 >= len(data) {
					continue
				}
				dst := ids[int(data[i])%len(ids)]
				old := ids[int(data[i+1])%len(ids)]
				repl := ids[int(data[i+2])%len(ids)]
				i += 3
				_ = n.PatchFanin(dst, old, ABCObjectEdge{ID: repl})
			}
			if err := n.Validate(); err != nil {
				t.Fatal(err)
			}
		}
	})
}
