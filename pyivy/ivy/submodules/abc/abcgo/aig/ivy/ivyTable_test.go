package ivy

import "testing"

func ivyTableObject(p *Ivy_Man_t, typ Ivy_Type_t, a, b *Ivy_Obj_t, init Ivy_Init_t) *Ivy_Obj_t {
	node := Ivy_ManFetchMemory(p)
	node.Id = int32(len(p.vObjs))
	node.Type = typ
	node.Init = init
	node.pFanin0, node.pFanin1 = a, b
	p.vObjs = append(p.vObjs, node)
	return node
}
func TestIvyTableIdentityAndPolarity(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	b := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	and := ivyTableObject(p, IVY_AND, a, b, IVY_INIT_NONE)
	Ivy_TableInsert(p, and)
	ghost := Ivy_ObjCreateGhost(p, a, b, IVY_AND, IVY_INIT_NONE)
	if Ivy_TableLookup(p, ghost) != and {
		t.Fatal("missing structural match")
	}
	if Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, Ivy_Not(a), b, IVY_AND, IVY_INIT_NONE)) != nil {
		t.Fatal("complement polarity not keyed")
	}
	xor := ivyTableObject(p, IVY_EXOR, a, b, IVY_INIT_NONE)
	Ivy_TableInsert(p, xor)
	if Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, a, b, IVY_EXOR, IVY_INIT_NONE)) != xor {
		t.Fatal("node type not keyed")
	}
	l0 := ivyTableObject(p, IVY_LATCH, a, nil, IVY_INIT_0)
	l1 := ivyTableObject(p, IVY_LATCH, a, nil, IVY_INIT_1)
	Ivy_TableInsert(p, l0)
	Ivy_TableInsert(p, l1)
	if Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, a, nil, IVY_LATCH, IVY_INIT_0)) != l0 || Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, a, nil, IVY_LATCH, IVY_INIT_1)) != l1 {
		t.Fatal("initial state not keyed")
	}
	Ivy_TableResize(p)
	if Ivy_TableCountEntries(p) != 4 || Ivy_TableLookup(p, and) != and {
		t.Fatal("resize lost entries")
	}
	Ivy_TableDelete(p, and)
	if Ivy_TableLookup(p, and) != nil || Ivy_TableCountEntries(p) != 3 {
		t.Fatal("deletion affected another key")
	}
}
func TestIvyTableTransfersObjectIdentity(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	b := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	old := ivyTableObject(p, IVY_AND, a, b, IVY_INIT_NONE)
	next := ivyTableObject(p, IVY_NONE, nil, nil, IVY_INIT_NONE)
	Ivy_TableInsert(p, old)
	Ivy_ObjOverwrite(next, old)
	Ivy_TableUpdate(p, old, next.Id)
	if Ivy_TableLookup(p, next) != next || Ivy_TableCountEntries(p) != 1 {
		t.Fatal("table ID transfer")
	}
	p.vObjs[old.Id] = nil
	Ivy_ManRecycleMemory(p, old)
	if Ivy_TableLookup(p, next) != next {
		t.Fatal("recycling old storage lost replacement")
	}
}
func TestIvyTableInvalidMutations(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	b := ivyTableObject(p, IVY_PI, nil, nil, IVY_INIT_NONE)
	node := ivyTableObject(p, IVY_AND, a, b, IVY_INIT_NONE)
	Ivy_TableInsert(p, node)
	mustPanic := func(f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("invalid mutation accepted")
			}
		}()
		f()
	}
	mustPanic(func() { Ivy_TableInsert(p, node) })
	mustPanic(func() { Ivy_TableUpdate(p, node, a.Id) })
	mustPanic(func() { Ivy_ManRecycleMemory(p, node) })
	Ivy_TableDelete(p, node)
	mustPanic(func() { Ivy_TableDelete(p, node) })
}
