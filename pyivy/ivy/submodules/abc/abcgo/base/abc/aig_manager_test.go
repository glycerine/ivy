package abc

import (
	"math/rand"
	"testing"
)

func TestAigManagerBooleanOperations(t *testing.T) {
	n := NewNetwork("aig")
	m := Abc_AigAlloc(n)
	a, b, c := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	and := Abc_AigAnd(m, a, b)
	if and != Abc_AigAnd(m, b, a) || Abc_AigAnd(m, a, a) != a.Signal() || Abc_AigAnd(m, a, Abc_ObjNot(a)) != n.One().Not() {
		t.Fatal("strash")
	}
	xor := Abc_AigXor(m, a, b)
	mux := Abc_AigMux(m, c, a, b)
	var typ int32
	if Abc_AigXorLookup(m, a, b, &typ).Object == nil || Abc_AigMuxLookup(m, c, a, b, &typ).Object == nil {
		t.Fatal("lookup")
	}
	for i := 0; i < 8; i++ {
		av, bv, cv := i&1 != 0, i&2 != 0, i&4 != 0
		for _, check := range []struct {
			e    Edge
			want bool
		}{{and, av && bv}, {xor, av != bv}, {mux, cv && av || !cv && bv}} {
			got, err := n.Evaluate(check.e, map[*Object]bool{a: av, b: bv, c: cv})
			if err != nil || got != check.want {
				t.Fatal(i, got, err)
			}
		}
	}
	Abc_AigResize(m)
	Abc_AigRehash(m)
	Abc_AigCheckFaninOrder(m)
	if Abc_AigCheck(m) != 1 {
		t.Fatal("check")
	}
}
func TestAigReplacementResimplifiesFanouts(t *testing.T) {
	n := NewNetwork("replacement")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	old := Abc_AigAnd(m, a, b)
	fanout := Abc_AigAnd(m, old, a)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, fanout.Not())
	oldID, fanoutID := old.Object.ID, fanout.Object.ID
	Abc_AigReplace(m, old.Object, a, 1)
	if po.fanins[0] != a.Signal().Not() {
		t.Fatal("fanout did not collapse", po.fanins)
	}
	if n.Object(oldID) != nil || n.Object(fanoutID) != nil {
		t.Fatal("dead cone retained")
	}
	if Abc_AigCheck(m) != 1 {
		t.Fatal("invalid strash after replacement")
	}
}
func TestAigReplacementPreservesChangedFanoutIdentity(t *testing.T) {
	n := NewNetwork("identity")
	m := Abc_AigAlloc(n)
	a, b, c, d := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	old := Abc_AigAnd(m, a, b)
	fanout := Abc_AigAnd(m, old, c)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, fanout)
	Abc_AigReplace(m, old.Object, d, 1)
	if po.fanins[0].Object != fanout.Object {
		t.Fatal("unnecessarily changed fanout identity")
	}
	for i := 0; i < 4; i++ {
		cv, dv := i&1 != 0, i&2 != 0
		got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: false, b: false, c: cv, d: dv})
		if err != nil || got != (cv && dv) {
			t.Fatal(got, err)
		}
	}
	if Abc_AigCheck(m) != 1 {
		t.Fatal("invalid manager")
	}
}
func TestAigMiterAndUpdateTracking(t *testing.T) {
	n := NewNetwork("tracking")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	var updated *[]*Object
	added := Abc_AigUpdateStart(m, &updated)
	pairs := []Edge{a.Signal(), b.Signal(), a.Signal(), a.Signal()}
	miter := Abc_AigMiter(m, &pairs, 0)
	if len(pairs) != 2 || len(*added) == 0 {
		t.Fatal("miter/tracking")
	}
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, miter)
	dangling := Abc_AigAnd(m, a, b)
	if len(dangling.Object.fanouts) != 0 {
		t.Fatal("expected dangling AND")
	}
	if removed := Abc_AigCleanup(m); removed < 1 || len(*updated) == 0 {
		t.Fatal("cleanup/tracking", removed)
	}
	Abc_AigUpdateReset(m)
	if len(*added) != 0 || len(*updated) != 0 {
		t.Fatal("tracking reset")
	}
	Abc_AigUpdateStop(m)
	got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: false, b: true})
	if err != nil || !got {
		t.Fatal(got, err)
	}
	if Abc_AigMiter2(m, nil) != n.One().Not() {
		t.Fatal("empty miter")
	}
}

func TestAigLookupAlternatePhases(t *testing.T) {
	n := NewNetwork("lookup-phases")
	m := Abc_AigAlloc(n)
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	xor := Abc_AigAnd(m, Abc_AigAnd(m, a.Signal().Not(), b.Signal().Not()).Not(), Abc_AigAnd(m, a, b).Not())
	mux := Abc_AigAnd(m, Abc_AigAnd(m, c, a.Signal().Not()).Not(), Abc_AigAnd(m, c.Signal().Not(), b.Signal().Not()).Not())
	var typ int32
	if got := Abc_AigXorLookup(m, a, b, &typ); got != xor || typ != 1 {
		t.Fatal(got, typ)
	}
	if got := Abc_AigMuxLookup(m, c, a, b, &typ); got != mux || typ != 1 {
		t.Fatal(got, typ)
	}
	Abc_AigSetNodePhases(n)
	if !n.constant.Phase || a.Phase || b.Phase || c.Phase || xor.Object.Phase || mux.Object.Phase {
		t.Fatal("zero-input phases")
	}
}
func TestAigRecreationRejectsForeignInputsBeforeMutation(t *testing.T) {
	n := NewNetwork("foreign-recreate")
	m := Abc_AigAlloc(n)
	a := n.NewObject(ObjectPI)
	empty := n.NewObject(ObjectNode)
	foreign := NewNetwork("foreign").NewObject(ObjectPI)
	defer func() {
		if recover() == nil {
			t.Error("accepted foreign operand")
		}
		if len(empty.fanins) != 0 || len(a.fanouts) != 0 {
			t.Error("failed recreation mutated network")
		}
	}()
	Abc_AigAndCreateFrom(m, a, foreign, empty)
}
func TestAigReplacementRejectsDependentTarget(t *testing.T) {
	n := NewNetwork("cyclic-replace")
	m := Abc_AigAlloc(n)
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	old := Abc_AigAnd(m, a, b)
	target := Abc_AigAnd(m, old, c)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, target)
	defer func() {
		if recover() == nil {
			t.Error("accepted dependent replacement")
		}
		if err := n.Check(); err != nil {
			t.Error("mutated invalid replacement", err)
		}
	}()
	Abc_AigReplace(m, old.Object, target, 1)
}

func TestAigReplacementRandomizedTruthTables(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	for trial := 0; trial < 80; trial++ {
		n := NewNetwork("random")
		m := Abc_AigAlloc(n)
		var inputs []*Object
		edges := []Edge{n.One()}
		for i := 0; i < 4; i++ {
			o := n.NewObject(ObjectPI)
			inputs = append(inputs, o)
			edges = append(edges, o.Signal())
		}
		for i := 0; i < 25; i++ {
			a, b := edges[rng.Intn(len(edges))], edges[rng.Intn(len(edges))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, Abc_AigAnd(m, a, b))
		}
		var candidates []*Object
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			if len(o.fanouts) > 0 {
				candidates = append(candidates, o)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		old := candidates[rng.Intn(len(candidates))]
		replacement := inputs[rng.Intn(len(inputs))].Signal()
		if rng.Intn(2) == 0 {
			replacement = replacement.Not()
		}
		var outputs []*Object
		for _, e := range edges[len(edges)-4:] {
			po := n.NewObject(ObjectPO)
			Abc_ObjAddFanin(po, e)
			outputs = append(outputs, po)
		}
		expected := make([][]bool, 16)
		for pattern := 0; pattern < 16; pattern++ {
			values := map[*Object]bool{n.constant: true}
			for i, o := range inputs {
				values[o] = pattern&(1<<i) != 0
			}
			replacementValue := values[replacement.Object] != replacement.Complement
			var eval func(*Object) bool
			eval = func(o *Object) bool {
				if o == old {
					return replacementValue
				}
				if value, ok := values[o]; ok {
					return value
				}
				var value bool
				if o.IsCO() {
					e := o.fanins[0]
					value = eval(e.Object) != e.Complement
				} else {
					a, b := o.fanins[0], o.fanins[1]
					value = (eval(a.Object) != a.Complement) && (eval(b.Object) != b.Complement)
				}
				values[o] = value
				return value
			}
			for _, po := range outputs {
				expected[pattern] = append(expected[pattern], eval(po))
			}
		}
		Abc_AigReplace(m, old, replacement, 1)
		if Abc_AigCheck(m) != 1 {
			t.Fatalf("trial %d: invalid graph", trial)
		}
		for pattern := 0; pattern < 16; pattern++ {
			values := map[*Object]bool{}
			for i, o := range inputs {
				values[o] = pattern&(1<<i) != 0
			}
			for i, po := range outputs {
				got, err := n.Evaluate(po.Signal(), values)
				if err != nil || got != expected[pattern][i] {
					t.Fatalf("trial %d input %d output %d: %v %v want %v", trial, pattern, i, got, err, expected[pattern][i])
				}
			}
		}
	}
}

func TestAigHashRemovalIsVisibleToLookup(t *testing.T) {
	n := NewNetwork("hash-delete")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	and := Abc_AigAnd(m, a, b)
	Abc_AigAndDelete(m, and.Object)
	if e := Abc_AigAndLookup(m, a, b); e.Object != nil {
		t.Fatal("deleted hash entry rediscovered by graph scan")
	}
}

func TestAigChoiceFanoutsAndCycles(t *testing.T) {
	n := NewNetwork("choices")
	m := Abc_AigAlloc(n)
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := Abc_AigAnd(m, a, b)
	alternative := Abc_AigAnd(m, a, c)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, root)
	root.Object.Choice = alternative.Object
	if Abc_AigNodeIsChoice(root.Object) != 1 || Abc_AigCheck(m) != 1 {
		t.Fatal("valid choice")
	}
	alternative.Object.Choice = root.Object
	if Abc_AigCheck(m) != 0 {
		t.Fatal("choice cycle accepted")
	}
	alternative.Object.Choice = nil
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, alternative)
	if Abc_AigCheck(m) != 0 {
		t.Fatal("choice alternative with fanout accepted")
	}
}
