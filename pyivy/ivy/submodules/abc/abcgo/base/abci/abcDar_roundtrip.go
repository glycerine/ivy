package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
)

func abcGraphMappedEdge(g *aig.Graph, e aig.Edge) abc.Edge {
	v := g.ObjectData(e)
	mapped, ok := v.(abc.Edge)
	if !ok || mapped.Object == nil {
		panic(fmt.Sprintf("abci: missing AIG-to-network copy for %d", e.NodeID()))
	}
	if e.Complemented() {
		mapped = mapped.Not()
	}
	return mapped
}

func abcGraphSetMapped(g *aig.Graph, e aig.Edge, mapped abc.Edge) {
	g.SetObjectData(e, mapped)
}

func darReachableNodes(g *aig.Graph) []aig.Edge {
	var nodes []aig.Edge
	for _, output := range g.Outputs() {
		aig.Aig_ManDfs_rec(g, output, &nodes)
	}
	return nodes
}

func darAllNodes(g *aig.Graph) []aig.Edge {
	var nodes []aig.Edge
	for id, node := range g.Nodes() {
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			aig.Aig_ManDfs_rec(g, aig.Edge(id)<<1, &nodes)
		}
	}
	return nodes
}

func abcNtkFromDarChoices(old *abc.Network, g *aig.Graph) *abc.Network {
	if old == nil || g == nil || aig.Aig_ManHasEquivs(g) == 0 || aig.Aig_ManBufNum(g) != 0 {
		return nil
	}
	n := abc.Abc_NtkStartFrom(old, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	n.Constraints = aig.Aig_ManConstrNum(g)
	n.Barriers = old.Barriers
	aig.Aig_ManCleanData(g)
	abcGraphSetMapped(g, aig.Aig_ManConst1(g), n.One())
	for i := int32(0); i < aig.Aig_ManCiNum(g); i++ {
		abcGraphSetMapped(g, aig.Aig_ManCi(g, i), abc.Abc_NtkCi(n, i).Signal())
	}
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	for _, e := range darAllNodes(g) {
		node := g.Node(e)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		left, right := abcGraphMappedEdge(g, node.Fanins[0]), abcGraphMappedEdge(g, node.Fanins[1])
		mapped := abc.Abc_AigAnd(manager, left, right)
		abcGraphSetMapped(g, e, mapped)
	}
	for _, e := range darAllNodes(g) {
		if equiv := aig.Aig_ObjEquiv(g, e); equiv != aig.InvalidEdge {
			abcGraphMappedEdge(g, e).Object.Choice = abcGraphMappedEdge(g, equiv).Object
		}
	}
	for i, output := range g.Outputs() {
		node := g.Node(output)
		if node.Kind != aig.Output {
			return nil
		}
		abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, int32(i)), abcGraphMappedEdge(g, node.Fanins[0]))
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkFromDar(): Network check has failed.")
	}
	return n
}

func abcNtkFromDarSeq(old *abc.Network, g *aig.Graph) *abc.Network {
	if old == nil || g == nil || old.Barriers != 0 {
		return nil
	}
	regs := aig.Aig_ManRegNum(g)
	if regs < 0 || int(regs) > len(g.Inputs()) || int(regs) > len(g.Outputs()) {
		return nil
	}
	n := abc.Abc_NtkStartFromNoLatches(old, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	n.Constraints = aig.Aig_ManConstrNum(g)
	n.Barriers = int(g.BarrierBuffers)
	abcGraphSetMapped(g, aig.Aig_ManConst1(g), n.One())
	piCount := aig.Aig_ManCiNum(g) - aig.Aig_ManRegNum(g)
	for i := int32(0); i < piCount; i++ {
		if i >= int32(len(n.Inputs())) {
			return nil
		}
		abcGraphSetMapped(g, aig.Aig_ManCi(g, i), abc.Abc_NtkPi(n, i).Signal())
	}
	// C creates one init-zero latch for every register, with LI/LO terminals
	// mapped back to the corresponding AIG CO/CI suffix entries.
	for i := int32(0); i < regs; i++ {
		latch := abc.Abc_NtkCreateLatch(n)
		li, lo := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)
		abc.Abc_ObjAddFanin(latch, li)
		abc.Abc_ObjAddFanin(lo, latch)
		abc.Abc_LatchSetInit0(latch)
		abcGraphSetMapped(g, aig.Aig_ManLi(g, i), li.Signal())
		abcGraphSetMapped(g, aig.Aig_ManLo(g, i), lo.Signal())
	}
	abc.Abc_NtkAddDummyBoxNames(n)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	for _, e := range darReachableNodes(g) {
		node := g.Node(e)
		if node.Kind == aig.Buffer {
			abcGraphSetMapped(g, e, abcGraphMappedEdge(g, node.Fanins[0]))
			continue
		}
		left := abcGraphMappedEdge(g, node.Fanins[0])
		right := abcGraphMappedEdge(g, node.Fanins[1])
		var mapped abc.Edge
		if node.Kind == aig.ExclusiveOr {
			mapped = abc.Abc_AigXor(manager, left, right)
		} else {
			mapped = abc.Abc_AigAnd(manager, left, right)
		}
		abcGraphSetMapped(g, e, mapped)
	}
	for i, output := range g.Outputs() {
		node := g.Node(output)
		if node.Kind != aig.Output {
			return nil
		}
		abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, int32(i)), abcGraphMappedEdge(g, node.Fanins[0]))
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkFromIvySeq(): Network check has failed.")
	}
	return n
}

func abcNtkAfterTrim(g *aig.Graph, old *abc.Network) *abc.Network {
	if g == nil || old == nil || g.Assertions != 0 || old.Barriers != 0 ||
		aig.Aig_ManRegNum(g) > abc.Abc_NtkLatchNum(old) ||
		saig.Saig_ManPiNum(g) > abc.Abc_NtkCiNum(old) ||
		saig.Saig_ManPoNum(g) != abc.Abc_NtkPoNum(old) ||
		g.InputNumbers == nil {
		return nil
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	n.Constraints = aig.Aig_ManConstrNum(g)
	n.Barriers = int(g.BarrierBuffers)
	abcGraphSetMapped(g, aig.Aig_ManConst1(g), n.One())
	regs := aig.Aig_ManRegNum(g)
	piCount := aig.Aig_ManCiNum(g) - regs
	poCount := aig.Aig_ManCoNum(g) - regs
	if piCount < 0 || poCount < 0 || len(g.InputNumbers) < int(aig.Aig_ManCiNum(g)) {
		return nil
	}
	for i := int32(0); i < piCount; i++ {
		obj := abc.Abc_NtkCreatePi(n)
		abcGraphSetMapped(g, aig.Aig_ManCi(g, i), obj.Signal())
		oldIndex := g.InputNumbers[i]
		if oldIndex < 0 || oldIndex >= abc.Abc_NtkCiNum(old) {
			return nil
		}
		abc.Abc_ObjAssignName(obj, abc.Abc_ObjName(abc.Abc_NtkCi(old, oldIndex)), "")
	}
	for i := int32(0); i < poCount; i++ {
		obj := abc.Abc_NtkCreatePo(n)
		oldObj := abc.Abc_NtkCo(old, i)
		abc.Abc_ObjAssignName(obj, abc.Abc_ObjName(oldObj), "")
	}
	for i := int32(0); i < regs; i++ {
		latch := abc.Abc_NtkCreateLatch(n)
		li, lo := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)
		abc.Abc_ObjAddFanin(latch, li)
		abc.Abc_ObjAddFanin(lo, latch)
		abc.Abc_LatchSetInit0(latch)
		abcGraphSetMapped(g, aig.Aig_ManLi(g, i), li.Signal())
		abcGraphSetMapped(g, aig.Aig_ManLo(g, i), lo.Signal())
		oldCi := g.InputNumbers[piCount+i]
		if oldCi < 0 || oldCi >= abc.Abc_NtkCiNum(old) {
			return nil
		}
		abc.Abc_ObjAssignName(lo, abc.Abc_ObjName(abc.Abc_NtkCi(old, oldCi)), "")
		abc.Abc_ObjAssignName(li, abc.Abc_ObjName(abc.Abc_NtkCo(old, poCount+i)), "")
	}
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	for _, e := range darReachableNodes(g) {
		node := g.Node(e)
		if node.Kind == aig.Buffer {
			abcGraphSetMapped(g, e, abcGraphMappedEdge(g, node.Fanins[0]))
		} else {
			abcGraphSetMapped(g, e, abc.Abc_AigAnd(manager,
				abcGraphMappedEdge(g, node.Fanins[0]), abcGraphMappedEdge(g, node.Fanins[1])))
		}
	}
	for i, output := range g.Outputs() {
		node := g.Node(output)
		abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, int32(i)), abcGraphMappedEdge(g, node.Fanins[0]))
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkAfterTrim(): Network check has failed.")
	}
	return n
}

func abcNtkFromCellMappedGia(p *gia.Gia_Man_t, useBuffs int32) *abc.Network {
	if p == nil || gia.Gia_ManHasCellMapping(p) == 0 {
		return nil
	}
	lib, _ := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
	if lib == nil {
		return nil
	}
	var cellCount int32
	cells := mio.Mio_CollectRootsNewDefault2(6, &cellCount, 0)
	if len(cells) == 0 {
		return nil
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_MAP, 1)
	n.Name = gia.Gia_ManName(p)
	n.Spec = gia.Gia_ManSpec(p)
	n.FunctionManager = lib
	copyLits := make([]int32, 2*gia.Gia_ManObjNum(p))
	for i := range copyLits {
		copyLits[i] = -1
	}
	write := func(id, phase, objID int32) { copyLits[int(2*id+phase)] = objID }
	read := func(id, phase int32) *abc.Object {
		idx := int(2*id + phase)
		if idx < 0 || idx >= len(copyLits) || copyLits[idx] < 0 {
			panic("abci: mapped GIA fanin has no cell copy")
		}
		return abc.Abc_NtkObj(n, copyLits[idx])
	}
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		write(gia.Gia_ObjId(p, gia.Gia_ManPi(p, i)), 0, abc.Abc_ObjId(abc.Abc_NtkCreatePi(n)))
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		write(gia.Gia_ObjId(p, gia.Gia_ManPo(p, i)), 0, abc.Abc_ObjId(abc.Abc_NtkCreatePo(n)))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		latch := abc.Abc_NtkCreateLatch(n)
		li, lo := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)
		abc.Abc_ObjAddFanin(latch, li)
		abc.Abc_ObjAddFanin(lo, latch)
		abc.Abc_LatchSetInit0(latch)
		write(gia.Gia_ObjId(p, gia.Gia_ManRi(p, i)), 0, abc.Abc_ObjId(li))
		write(gia.Gia_ObjId(p, gia.Gia_ManRo(p, i)), 0, abc.Abc_ObjId(lo))
	}
	for coIndex := int32(0); coIndex < gia.Gia_ManCoNum(p); coIndex++ {
		co := gia.Gia_ManCo(p, coIndex)
		if gia.Gia_ObjFaninId0p(p, co) == 0 {
			phase := gia.Gia_ObjFaninC0(co)
			key := int(phase)
			if copyLits[key] == -1 {
				var c *abc.Object
				if phase == 0 {
					c = abc.Abc_NtkCreateNodeConst0(n)
				} else {
					c = abc.Abc_NtkCreateNodeConst1(n)
				}
				write(0, phase, abc.Abc_ObjId(c))
			}
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsBuf(obj) != 0 && gia.Gia_ObjFaninId0p(p, obj) == 0 {
			phase := gia.Gia_ObjFaninC0(obj)
			if copyLits[int(phase)] == -1 {
				var c *abc.Object
				if phase == 0 {
					c = abc.Abc_NtkCreateNodeConst0(n)
				} else {
					c = abc.Abc_NtkCreateNodeConst1(n)
				}
				write(0, phase, abc.Abc_ObjId(c))
			}
		}
	}
	for i := int32(2); i < 2*gia.Gia_ManObjNum(p); i++ {
		if gia.Gia_ObjIsCell(p, i) == 0 {
			continue
		}
		id, phase := i>>1, i&1
		var node *abc.Object
		skip := int32(0)
		switch {
		case gia.Gia_ObjIsCellBuf(p, i) != 0:
			node = abc.Abc_NtkCreateNode(n)
			fan := gia.Gia_ObjFaninLit0p(p, gia.Gia_ManObj(p, id))
			abc.Abc_ObjAddFanin(node, read(fan>>1, fan&1))
			node.Barrier = true
			abc.Abc_NtkIncrementBarBufs2(n)
		case gia.Gia_ObjIsCellInv(p, i) != 0:
			invLit := i ^ 1
			if phase == 0 {
				if gia.Gia_ObjCellId(p, invLit) < 0 {
					return nil
				}
				node = abc.Abc_NtkCreateNode(n)
				for _, fan := range gia.Gia_ObjCellFanins(p, invLit) {
					abc.Abc_ObjAddFanin(node, read(fan>>1, fan&1))
				}
				cellID := gia.Gia_ObjCellId(p, invLit)
				if int(cellID) >= len(cells) {
					return nil
				}
				node.Data = mio.Mio_LibraryReadGateByName(lib, cells[cellID].Name(), "")
				write(id, invLit&1, abc.Abc_ObjId(node))
				skip = 1
			}
			node = abc.Abc_NtkCreateNode(n)
			abc.Abc_ObjAddFanin(node, read(id, invLit&1))
			if len(cells) <= 3 {
				return nil
			}
			node.Data = mio.Mio_LibraryReadGateByName(lib, cells[3].Name(), "")
		default:
			cellID := gia.Gia_ObjCellId(p, i)
			if cellID < 0 || int(cellID) >= len(cells) {
				return nil
			}
			node = abc.Abc_NtkCreateNode(n)
			for _, fan := range gia.Gia_ObjCellFanins(p, i) {
				abc.Abc_ObjAddFanin(node, read(fan>>1, fan&1))
			}
			node.Data = mio.Mio_LibraryReadGateByName(lib, cells[cellID].Name(), "")
		}
		write(id, phase, abc.Abc_ObjId(node))
		i += skip
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		co := gia.Gia_ManCo(p, i)
		driver := read(gia.Gia_ObjFaninId0p(p, co), gia.Gia_ObjFaninC0(co))
		abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, i), driver)
	}
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	abc.Abc_NtkAddDummyBoxNames(n)
	abc.Abc_NtkLogicMakeSimpleCos(n, 1-useBuffs)
	if gia.Gia_ManPiNum(p) != abc.Abc_NtkPiNum(n) || gia.Gia_ManPoNum(p) != abc.Abc_NtkPoNum(n) || gia.Gia_ManRegNum(p) != abc.Abc_NtkLatchNum(n) {
		return nil
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkFromMappedGia(): Network check has failed.")
	}
	return n
}
