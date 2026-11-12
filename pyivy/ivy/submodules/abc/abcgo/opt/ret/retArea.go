package ret

import "github.com/glycerine/abcgo/base/abc"

func Abc_NtkRetimeMinArea(n *abc.Network, forwardOnly, backwardOnly, useOldNames, verbose int32) int32 {
	if forwardOnly != 0 && backwardOnly != 0 {
		panic("ret: conflicting retiming directions")
	}
	if abc.Abc_NtkIsSopLogic(n) == 0 || abc.Abc_NtkBoxNum(n) != abc.Abc_NtkLatchNum(n) {
		panic("ret: minimum area needs SOP logic without other boxes")
	}
	before := len(retLatches(n))
	abc.Abc_NtkOrderCisCos(n)
	if backwardOnly == 0 {
		for {
			_, changed := Abc_NtkRetimeMinAreaOne(n, 1, useOldNames, verbose)
			if !changed {
				break
			}
		}
	}
	values := Abc_NtkRetimeCollectLatchValues(n)
	var total *abc.Network
	if forwardOnly == 0 {
		for {
			bottom, changed := Abc_NtkRetimeMinAreaOne(n, 0, useOldNames, verbose)
			if !changed {
				break
			}
			total = retAttachBottom(total, bottom)
		}
	}
	Abc_NtkRetimeInsertLatchValues(n, Abc_NtkRetimeInitialValues(total, values, verbose))
	return int32(before - len(retLatches(n)))
}

// The boolean replaces the original C pointer value 1 used only as a forward
// progress sentinel. A real network is returned for backward state constraints.
func Abc_NtkRetimeMinAreaOne(n *abc.Network, forward, useOldNames, verbose int32) (*abc.Network, bool) {
	Abc_NtkRetimeMinAreaPrepare(n, forward)
	cut := Abc_NtkMaxFlow(n, forward, verbose)
	defer func() {
		for _, o := range n.Objects() {
			o.MarkA = false
		}
	}()
	if len(cut) >= len(retLatches(n)) {
		return nil, false
	}
	var cone *abc.Network
	if forward != 0 {
		Abc_NtkRetimeMinAreaInitValues(n, cut)
	} else {
		cone = Abc_NtkRetimeMinAreaConstructNtk(n, cut)
	}
	Abc_NtkRetimeMinAreaUpdateLatches(n, cut, forward, useOldNames)
	return cone, true
}
func Abc_NtkMarkCone_rec(o *abc.Object, forward int32) {
	if o.MarkA {
		return
	}
	o.MarkA = true
	if forward != 0 {
		for _, next := range o.Fanouts() {
			Abc_NtkMarkCone_rec(next, forward)
		}
	} else {
		for _, e := range o.Fanins() {
			Abc_NtkMarkCone_rec(e.Object, forward)
		}
	}
}
func Abc_NtkUnmarkCone_rec(o *abc.Object, forward int32) {
	if !o.MarkA || o.Kind == abc.ObjectLatch {
		return
	}
	o.MarkA = false
	if forward != 0 {
		for _, next := range o.Fanouts() {
			Abc_NtkUnmarkCone_rec(next, forward)
		}
	} else {
		for _, e := range o.Fanins() {
			Abc_NtkUnmarkCone_rec(e.Object, forward)
		}
	}
}
func Abc_NtkRetimeMinAreaPrepare(n *abc.Network, forward int32) {
	if forward != 0 {
		for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
			o.MarkA = true
		}
		for _, o := range retLatches(n) {
			o.MarkA = true
			retFan0(o).MarkA = true
		}
		for _, o := range n.ObjectsOfKind(abc.ObjectPI) {
			Abc_NtkMarkCone_rec(o, forward)
		}
		var boundary []*abc.Object
		for _, o := range n.Objects() {
			if o.MarkA {
				for _, e := range o.Fanins() {
					if !e.Object.MarkA {
						boundary = append(boundary, e.Object)
					}
				}
			}
		}
		for _, o := range boundary {
			o.MarkA = true
		}
	} else {
		for _, o := range n.ObjectsOfKind(abc.ObjectPI) {
			o.MarkA = true
		}
		for _, o := range retLatches(n) {
			o.MarkA = true
			o.Fanouts()[0].MarkA = true
		}
		for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
			Abc_NtkMarkCone_rec(o, forward)
		}
	}
}
func Abc_NtkRetimeMinAreaInitValues_rec(o *abc.Object) int32 {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return int32(o.Temp)
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	if o.Kind == abc.ObjectBO {
		o.Temp = int(Abc_NtkRetimeMinAreaInitValues_rec(retFan0(o)))
		return int32(o.Temp)
	}
	if o.Kind == abc.ObjectConst1 {
		o.Temp = 1
		return 1
	}
	if o.Kind != abc.ObjectNode {
		panic("ret: invalid forward initial-state cone")
	}
	for _, e := range o.Fanins() {
		Abc_NtkRetimeMinAreaInitValues_rec(e.Object)
	}
	o.Temp = int(Abc_ObjSopSimulate(o))
	return int32(o.Temp)
}
func Abc_NtkRetimeMinAreaInitValues(n *abc.Network, cut []*abc.Object) {
	abc.Abc_NtkIncrementTravId(n)
	for _, o := range retLatches(n) {
		o.Temp = int(retLatchValue(o))
		abc.Abc_NodeSetTravIdCurrent(o)
	}
	for _, o := range cut {
		Abc_NtkRetimeMinAreaInitValues_rec(o)
	}
	for _, o := range retLatches(n) {
		abc.Abc_NodeSetTravIdPrevious(o)
	}
}
func Abc_NtkRetimeMinAreaConstructNtk_rec(dst *abc.Network, o *abc.Object) *abc.Object {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return o.Copy
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	if o.Kind == abc.ObjectBI {
		o.Copy = Abc_NtkRetimeMinAreaConstructNtk_rec(dst, retFan0(o))
		return o.Copy
	}
	if o.Kind == abc.ObjectConst1 {
		o.Copy = dst.One().Object
		return o.Copy
	}
	if o.Kind != abc.ObjectNode {
		panic("ret: invalid backward initial-state cone")
	}
	for _, e := range o.Fanins() {
		Abc_NtkRetimeMinAreaConstructNtk_rec(dst, e.Object)
	}
	o.Copy = retSopNode(dst, o)
	for _, e := range o.Fanins() {
		copy := e.Object.Copy.Signal()
		if e.Complement {
			copy = copy.Not()
		}
		retAdd(o.Copy, copy)
	}
	return o.Copy
}
func Abc_NtkRetimeMinAreaConstructNtk(n *abc.Network, cut []*abc.Object) *abc.Network {
	dst := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	abc.Abc_NtkIncrementTravId(n)
	for _, o := range cut {
		o.Copy = abc.Abc_NtkCreatePi(dst)
		abc.Abc_NodeSetTravIdCurrent(o)
	}
	for _, o := range retLatches(n) {
		copy := Abc_NtkRetimeMinAreaConstructNtk_rec(dst, retFan0(o))
		retAdd(abc.Abc_NtkCreatePo(dst), copy.Signal())
		abc.Abc_NodeSetTravIdPrevious(o)
	}
	for _, o := range cut {
		abc.Abc_NodeSetTravIdPrevious(o)
	}
	return dst
}
func retPatchReferences(o, old, replacement *abc.Object) {
	for i, e := range o.Fanins() {
		if e.Object == old {
			r := replacement.Signal()
			r.Complement = e.Complement
			retPatch(o, i, r)
		}
	}
}
func Abc_NtkRetimeMinAreaUpdateLatches(n *abc.Network, cut []*abc.Object, forward, useOldNames int32) {
	old := retLatches(n)
	reused := make(map[*abc.Object]bool)
	var buffers []*abc.Object
	var order []int
	for _, box := range n.Objects() {
		if box.Kind == abc.ObjectWhitebox || box.Kind == abc.ObjectBlackbox {
			order = append(order, box.ID)
		}
	}
	for _, o := range cut {
		var latch, in, out *abc.Object
		if o.IsCI() && forward != 0 {
			out = o
			latch = retFan0(out)
			in = retFan0(latch)
			reused[latch] = true
			marked := false
			for _, next := range o.Fanouts() {
				marked = marked || next.MarkA
			}
			if marked {
				buf := abc.Abc_NtkCreateNodeBuf(n, retFan0(in))
				buf.SetName(o.Name() + "_buf")
				retPatch(in, 0, buf.Signal())
				buffers = append(buffers, buf)
				for _, next := range o.Fanouts() {
					if next.MarkA {
						retPatchReferences(next, o, buf)
					}
				}
			}
		} else if o.IsCO() && forward == 0 {
			in = o
			latch = in.Fanouts()[0]
			out = latch.Fanouts()[0]
			reused[latch] = true
		} else {
			out = abc.Abc_NtkCreateBo(n)
			latch = abc.Abc_NtkCreateLatch(n)
			in = abc.Abc_NtkCreateBi(n)
			base := o.Name()
			if useOldNames != 0 {
				base = latch.Name()
			}
			out.SetName(base + "_out")
			in.SetName(base + "_in")
			retAdd(out, latch.Signal())
			retAdd(latch, in.Signal())
			if forward != 0 {
				if o.Temp != 0 {
					abc.Abc_LatchSetInit1(latch)
				} else {
					abc.Abc_LatchSetInit0(latch)
				}
			}
			for _, next := range o.Fanouts() {
				if forward != 0 && !next.MarkA || forward == 0 && next.MarkA {
					retPatchReferences(next, o, out)
				}
			}
			retAdd(in, o.Signal())
		}
		order = append(order, latch.ID)
	}
	for _, buf := range buffers {
		retTransfer(buf, retFan0(buf))
		retDelete(buf)
	}
	for _, latch := range old {
		if reused[latch] {
			continue
		}
		out, in := latch.Fanouts()[0], retFan0(latch)
		retTransfer(out, retFan0(in))
		retDelete(out)
		retDelete(latch)
		retDelete(in)
	}
	if err := n.SetBoxOrder(order); err != nil {
		panic(err)
	}
	abc.Abc_NtkOrderCisCos(n)
}

// Compose state-transition constraints by substituting the new cone's outputs
// into the old cone's inputs. The resulting PIs follow the newest latch order.
func retAttachBottom(top, bottom *abc.Network) *abc.Network {
	if top == nil {
		return bottom
	}
	dst := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	copies := make(map[*abc.Object]abc.Edge)
	copies[bottom.One().Object] = dst.One()
	for _, o := range bottom.ObjectsOfKind(abc.ObjectPI) {
		copies[o] = abc.Abc_NtkCreatePi(dst).Signal()
	}
	copyLogic := func(src *abc.Network) {
		for _, o := range abc.Abc_NtkDfs(src, 0) {
			node := retSopNode(dst, o)
			for _, e := range o.Fanins() {
				v := copies[e.Object]
				if e.Complement {
					v = v.Not()
				}
				retAdd(node, v)
			}
			copies[o] = node.Signal()
		}
	}
	copyLogic(bottom)
	outputs, inputs := bottom.ObjectsOfKind(abc.ObjectPO), top.ObjectsOfKind(abc.ObjectPI)
	if len(outputs) != len(inputs) {
		panic("ret: initial-state cone interfaces differ")
	}
	for i, o := range inputs {
		e := outputs[i].Fanins()[0]
		v := copies[e.Object]
		if e.Complement {
			v = v.Not()
		}
		copies[o] = v
	}
	copies[top.One().Object] = dst.One()
	copyLogic(top)
	for _, o := range top.ObjectsOfKind(abc.ObjectPO) {
		e := o.Fanins()[0]
		v := copies[e.Object]
		if e.Complement {
			v = v.Not()
		}
		retAdd(abc.Abc_NtkCreatePo(dst), v)
	}
	return dst
}
