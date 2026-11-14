package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Abc_NtkMulti converts a strashed AIG into a BDD-function logic network.
// Boundary selection follows abcMulti.c; each output is rebuilt from its
// boundary cut, sharing nodes that are already represented in the output DAG.
func Abc_NtkMulti(n *Abc_Ntk_t, threshold, faninMax, cnf, multi, simple, factor int32) *Abc_Ntk_t {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 || threshold < 0 || faninMax <= 1 {
		return nil
	}
	if abc.Abc_NtkGetChoiceNum(n) > 0 {
		fmt.Println("Warning: The choice nodes in the AIG are removed by renoding.")
	}
	switch {
	case cnf != 0:
		Abc_NtkMultiSetBoundsCnf(n)
	case multi != 0:
		Abc_NtkMultiSetBoundsMulti(n, threshold)
	case simple != 0:
		Abc_NtkMultiSetBoundsSimple(n)
	case factor != 0:
		Abc_NtkMultiSetBoundsFactor(n)
	default:
		Abc_NtkMultiSetBounds(n, threshold, faninMax)
	}
	out := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD)
	if out == nil {
		Abc_NtkMultiCleanMarks(n)
		return nil
	}
	Abc_NtkMultiInt(n, out)
	abc.Abc_NtkFinalize(n, out)
	abc.Abc_NtkCleanCopy(n)
	abc.Abc_NtkMinimumBase(out)
	abc.Abc_NtkLogicMakeSimpleCos(out, 0)
	if n.Exdc != nil {
		out.Exdc = abc.Abc_NtkDup(n.Exdc)
	}
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkMulti: The network check has failed.")
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

// Abc_NtkMultiInt reconstructs each CO cone. The pCopy scratch on old nodes
// retains sharing across outputs, matching the original traversal.
func Abc_NtkMultiInt(n, out *Abc_Ntk_t) {
	dd, _ := out.FunctionManager.(*core.DdManager)
	if dd == nil {
		return
	}
	constant := abc.Abc_AigConst1(n).Object
	if constant != nil && abc.Abc_ObjFanoutNum(constant) > 0 {
		constant.Copy = abc.Abc_NtkCreateNode(out)
		constant.CopyEdge = constant.Copy.Signal()
		constant.Copy.Data = core.Cudd_ReadOne(dd)
		core.Cudd_Ref(constant.Copy.Data.(*core.DdNode))
	}
	for _, co := range n.Objects() {
		if !co.IsCO() || len(co.Fanins()) == 0 {
			continue
		}
		root := abc.Abc_ObjFanin0(co)
		if root.IsCI() {
			continue
		}
		Abc_NtkMultiRec(out, root)
	}
	Abc_NtkMultiCleanMarks(n)
}

func Abc_NtkMultiRec(out *Abc_Ntk_t, old *Abc_Obj_t) *Abc_Obj_t {
	if old == nil || old.Copy != nil {
		return old.Copy
	}
	if old.Kind != abc.ObjectNode || old.MarkA == false {
		return old
	}
	cone := Abc_NtkMultiCone(old)
	if len(cone) == 0 {
		return nil
	}
	fresh := abc.Abc_NtkCreateNode(out)
	for _, leaf := range cone {
		abc.Abc_ObjAddFanin(fresh, Abc_NtkMultiRec(out, leaf))
	}
	dd := out.FunctionManager.(*core.DdManager)
	fresh.Data = Abc_NtkMultiDeriveBdd(dd, old, cone)
	core.Cudd_Ref(fresh.Data.(*core.DdNode))
	old.Copy = fresh
	return fresh
}

func Abc_NtkMultiDeriveBdd(dd *core.DdManager, root *Abc_Obj_t, leaves []*Abc_Obj_t) *core.DdNode {
	for i, leaf := range leaves {
		leaf.CopyData = core.Cudd_bddIthVar(dd, i)
		core.Cudd_Ref(leaf.CopyData.(*core.DdNode))
		leaf.MarkC = true
	}
	result := Abc_NtkMultiDeriveBddRec(dd, root)
	core.Cudd_Ref(result)
	for _, leaf := range leaves {
		core.Cudd_RecursiveDeref(dd, leaf.CopyData.(*core.DdNode))
		leaf.CopyData = nil
		leaf.MarkC = false
	}
	for _, o := range root.Network().Objects() {
		if o.Kind == abc.ObjectNode {
			if scratch, ok := o.CopyData.(*core.DdNode); ok {
				core.Cudd_RecursiveDeref(dd, scratch)
			}
			o.CopyData = nil
			o.MarkC = false
		}
	}
	core.Cudd_Deref(result)
	return result
}

func Abc_NtkMultiDeriveBddRec(dd *core.DdManager, node *Abc_Obj_t) *core.DdNode {
	if node.MarkC {
		if b, ok := node.CopyData.(*core.DdNode); ok {
			return b
		}
		return core.Cudd_ReadOne(dd)
	}
	node.MarkC = true
	if node.Kind == abc.ObjectConst1 {
		node.CopyData = core.Cudd_ReadOne(dd)
		return node.CopyData.(*core.DdNode)
	}
	f0 := Abc_NtkMultiDeriveBddRec(dd, abc.Abc_ObjFanin0(node))
	core.Cudd_Ref(f0)
	f1 := Abc_NtkMultiDeriveBddRec(dd, abc.Abc_ObjFanin1(node))
	core.Cudd_Ref(f1)
	f0 = core.Cudd_NotCond(f0, abc.Abc_ObjFaninC0(node) != 0)
	f1 = core.Cudd_NotCond(f1, abc.Abc_ObjFaninC1(node) != 0)
	result := core.Cudd_bddAnd(dd, f0, f1)
	core.Cudd_Ref(result)
	core.Cudd_RecursiveDeref(dd, f0)
	core.Cudd_RecursiveDeref(dd, f1)
	node.CopyData = result
	return result
}

func Abc_NtkMultiLimitRec(node *Abc_Obj_t, cone []*Abc_Obj_t, max int, canStop, first bool) ([]*Abc_Obj_t, bool) {
	if !first && (node.MarkA || node.Kind != abc.ObjectNode) {
		return appendUniqueNode(cone, node), false
	}
	if !canStop {
		cone, _ = Abc_NtkMultiLimitRec(abc.Abc_ObjFanin0(node), cone, max, false, false)
		cone, _ = Abc_NtkMultiLimitRec(abc.Abc_ObjFanin1(node), cone, max, false, false)
		return cone, false
	}
	left, changed := Abc_NtkMultiLimitRec(abc.Abc_ObjFanin0(node), nil, max, true, false)
	if changed {
		return left, true
	}
	leftN := len(left)
	both, _ := Abc_NtkMultiLimitRec(abc.Abc_ObjFanin1(node), left, max, false, false)
	if len(both) <= max {
		return both, false
	}
	right, _ := Abc_NtkMultiLimitRec(abc.Abc_ObjFanin1(node), nil, max, false, false)
	if len(right) > max {
		_, changed = Abc_NtkMultiLimitRec(abc.Abc_ObjFanin1(node), nil, max, true, false)
		return nil, changed
	}
	if leftN >= len(right) {
		abc.Abc_ObjFanin0(node).MarkA = true
	} else {
		abc.Abc_ObjFanin1(node).MarkA = true
	}
	return nil, true
}

func Abc_NtkMultiLimit(node *Abc_Obj_t, max int) bool {
	_, changed := Abc_NtkMultiLimitRec(node, nil, max, true, true)
	return changed
}

func Abc_NtkMultiSetBounds(n *Abc_Ntk_t, threshold, max int32) {
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if (abc.Abc_ObjFanoutNum(node)-1)*abc.Abc_NodeMffcSize(node) > threshold {
			node.MarkA = true
		}
	}
	for _, co := range n.Objects() {
		if co.IsCO() && len(co.Fanins()) > 0 {
			abc.Abc_ObjFanin0(co).MarkA = true
		}
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if node.MarkA {
			for Abc_NtkMultiLimit(node, int(max)) {
			}
		}
	}
}

func Abc_NtkMultiSetBoundsCnf(n *Abc_Ntk_t) {
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if abc.Abc_ObjFanoutNum(node) > 1 {
			node.MarkA = true
		}
		if abc.Abc_NodeIsMuxType(node) != 0 {
			node.MarkA = true
			for _, i := range []int32{0, 1} {
				for _, j := range []int32{0, 1} {
					abc.Abc_ObjFanin(abc.Abc_ObjFanin(node, i), j).MarkA = true
				}
			}
		} else {
			for i := int32(0); i < 2; i++ {
				if abc.Abc_ObjFaninC(node, i) != 0 {
					abc.Abc_ObjFanin(node, i).MarkA = true
				}
			}
		}
	}
	markCoDrivers(n)
}

func Abc_NtkMultiSetBoundsMulti(n *Abc_Ntk_t, threshold int32) {
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if (abc.Abc_ObjFanoutNum(node)-1)*abc.Abc_NodeMffcSizeStop(node) > threshold {
			node.MarkA = true
		}
		for i := int32(0); i < 2; i++ {
			if abc.Abc_ObjFaninC(node, i) != 0 {
				abc.Abc_ObjFanin(node, i).MarkA = true
			}
		}
	}
	markCoDrivers(n)
}

func Abc_NtkMultiSetBoundsSimple(n *Abc_Ntk_t) {
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		node.MarkA = true
	}
}
func Abc_NtkMultiSetBoundsFactor(n *Abc_Ntk_t) {
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		node.MarkA = abc.Abc_ObjFanoutNum(node) > 1 && abc.Abc_NodeIsMuxControlType(node) == 0
	}
	markCoDrivers(n)
}
func markCoDrivers(n *Abc_Ntk_t) {
	for _, co := range n.Objects() {
		if co.IsCO() && len(co.Fanins()) > 0 {
			abc.Abc_ObjFanin0(co).MarkA = true
		}
	}
}

func Abc_NtkMultiCone(root *Abc_Obj_t) []*Abc_Obj_t {
	var cone []*Abc_Obj_t
	var visit func(*Abc_Obj_t)
	visit = func(node *Abc_Obj_t) {
		if node.MarkA || node.Kind != abc.ObjectNode {
			cone = appendUniqueNode(cone, node)
			return
		}
		visit(abc.Abc_ObjFanin0(node))
		visit(abc.Abc_ObjFanin1(node))
	}
	visit(abc.Abc_ObjFanin0(root))
	visit(abc.Abc_ObjFanin1(root))
	return cone
}
func appendUniqueNode(nodes []*Abc_Obj_t, node *Abc_Obj_t) []*Abc_Obj_t {
	for _, o := range nodes {
		if o == node {
			return nodes
		}
	}
	return append(nodes, node)
}
func Abc_NtkMultiCleanMarks(n *Abc_Ntk_t) {
	for _, node := range n.Objects() {
		node.MarkA = false
		node.MarkC = false
		node.CopyData = nil
	}
}
