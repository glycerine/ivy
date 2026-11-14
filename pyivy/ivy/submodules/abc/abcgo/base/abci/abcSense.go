package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fraig"
)

// sensitivityCopyEdge maps an AIG child through the copy edge created for its
// regular object, then reapplies the source edge's complement bit.
func sensitivityCopyEdge(source abc.Edge) abc.Edge {
	if source.Object == nil {
		panic("abci: sensitivity edge has no source object")
	}
	copy := abc.Abc_ObjCopyEdge(source.Object)
	if copy.Object == nil {
		panic("abci: sensitivity miter encountered an unmapped fanin")
	}
	if source.Complement {
		copy = copy.Not()
	}
	return copy
}

func sensitivityDataEdge(source abc.Edge) abc.Edge {
	if source.Object == nil {
		panic("abci: sensitivity data edge has no source object")
	}
	data, ok := source.Object.CopyData.(abc.Edge)
	if !ok || data.Object == nil {
		panic("abci: sensitivity miter encountered an unmapped data fanin")
	}
	if source.Complement {
		data = data.Not()
	}
	return data
}

// Abc_NtkSensitivityMiter_rec copies the transitive fanin of an AIG node into
// the sensitivity miter. The returned edge preserves complement phase.
func Abc_NtkSensitivityMiter_rec(miter *Abc_Ntk_t, node *Abc_Obj_t) abc.Edge {
	if node == nil || abc.Abc_ObjIsComplement(node) != 0 {
		panic("Abc_NtkSensitivityMiter_rec requires a regular AIG node")
	}
	if existing := abc.Abc_ObjCopyEdge(node); existing.Object != nil {
		return existing
	}
	first, second := abc.Abc_ObjChild0(node), abc.Abc_ObjChild1(node)
	Abc_NtkSensitivityMiter_rec(miter, first.Object)
	Abc_NtkSensitivityMiter_rec(miter, second.Object)
	man, ok := miter.FunctionManager.(*abc.Abc_Aig_t)
	if !ok || man == nil {
		panic("abci: sensitivity miter has no AIG manager")
	}
	copy := abc.Abc_AigAnd(man, sensitivityCopyEdge(first), sensitivityCopyEdge(second))
	node.SetCopyEdge(copy)
	return copy
}

func Abc_NtkSensitivityMiter(n *Abc_Ntk_t, input int32) *Abc_Ntk_t {
	if abc.Abc_NtkIsStrash(n) == 0 || input < 0 || input >= abc.Abc_NtkCiNum(n) {
		panic("Abc_NtkSensitivityMiter requires a strashed network and valid CI")
	}
	miter := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	miter.Name, miter.Spec = n.Name, n.Spec
	abc.Abc_NtkCleanCopy(n)
	constant := abc.Abc_AigConst1(miter)
	abc.Abc_AigConst1(n).Object.SetCopyEdge(constant)
	abc.Abc_ObjSetCopyData(abc.Abc_AigConst1(n).Object, constant)

	for i := int32(0); i < abc.Abc_NtkCiNum(n); i++ {
		ci := abc.Abc_NtkCi(n, i)
		copy := abc.Abc_NtkCreatePi(miter).Signal()
		ci.SetCopyEdge(copy)
		ci.CopyData = copy
	}
	selected := abc.Abc_NtkCi(n, input)
	selected.SetCopyEdge(constant.Not())
	selected.CopyData = constant
	abc.Abc_NtkAddDummyPiNames(miter)

	// Reverse DFS returns one linked list head per level. In this API, Copy is
	// the DFS list link; copy mappings are kept in CopyEdge independently.
	levels := abc.Abc_NtkDfsReverseNodes(n, []*Abc_Obj_t{selected}, 1)
	manager, ok := miter.FunctionManager.(*abc.Abc_Aig_t)
	if !ok || manager == nil {
		panic("abci: sensitivity miter has no AIG manager")
	}
	for _, head := range levels {
		for node := head; node != nil; {
			next := node.Copy // save the DFS link before replacing the copy edge
			if abc.Abc_ObjFaninNum(node) != 2 {
				panic(fmt.Sprintf("abci: reverse DFS node %d kind %d has %d fanins", node.ID, node.Kind, abc.Abc_ObjFaninNum(node)))
			}
			for _, child := range []abc.Edge{abc.Abc_ObjChild0(node), abc.Abc_ObjChild1(node)} {
				if abc.Abc_NodeIsTravIdCurrent(child.Object) == 0 {
					child.Object.CopyData = Abc_NtkSensitivityMiter_rec(miter, child.Object)
				}
			}
			copy := abc.Abc_AigAnd(manager, sensitivityCopyEdge(abc.Abc_ObjChild0(node)), sensitivityCopyEdge(abc.Abc_ObjChild1(node)))
			data := abc.Abc_AigAnd(manager, sensitivityDataEdge(abc.Abc_ObjChild0(node)), sensitivityDataEdge(abc.Abc_ObjChild1(node)))
			node.SetCopyEdge(copy)
			node.CopyData = data
			node = next
		}
	}

	output := constant.Not()
	for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
		co := abc.Abc_NtkCo(n, i)
		if abc.Abc_NodeIsTravIdCurrent(co) == 0 {
			continue
		}
		child := abc.Abc_ObjChild0(co)
		if i == abc.Abc_NtkCoNum(n)-1 {
			output = abc.Abc_AigAnd(manager, output, sensitivityDataEdge(child))
			output = abc.Abc_AigAnd(manager, output, sensitivityCopyEdge(child))
		} else {
			difference := abc.Abc_AigXor(manager, sensitivityCopyEdge(child), sensitivityDataEdge(child))
			output = abc.Abc_AigOr(manager, output, difference)
		}
	}
	po := abc.Abc_NtkCreatePo(miter)
	abc.Abc_ObjAddFanin(po, output)
	abc.Abc_ObjAssignName(po, "miter", "")
	if abc.Abc_NtkCheck(miter) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkSensitivityMiter: The network check has failed.")
		abc.Abc_NtkDelete(miter)
		return nil
	}
	return miter
}

func Abc_NtkSensitivity(n *Abc_Ntk_t, conflictLimit, verbose int32) *Vec_Int_t {
	if abc.Abc_NtkIsStrash(n) == 0 || abc.Abc_NtkLatchNum(n) != 0 {
		panic("Abc_NtkSensitivity requires a combinational strashed network")
	}
	params := fraig.Prove_Params_t{}
	fraig.Prove_ParamsSetDefault(&params)
	fraig.Prove_ParamsSetItersMax(&params, 3)
	fraig.Prove_ParamsSetMiteringLimitLast(&params, conflictLimit)
	result := vec.Vec_IntAlloc(100)
	progress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkCiNum(n))
	defer extra.Extra_ProgressBarStop(progress)
	for i := int32(0); i < abc.Abc_NtkCiNum(n); i++ {
		extra.Extra_ProgressBarUpdate(progress, i, "")
		miter := Abc_NtkSensitivityMiter(n, i)
		if miter == nil {
			panic("abci: sensitivity miter validation failed")
		}
		ret := Abc_NtkIvyProve(&miter, &params)
		if ret == -1 {
			vec.Vec_IntPush(result, i)
		} else if ret == 0 {
			pattern := Abc_NtkVerifySimulatePatternSlice(miter, miter.Model)
			if len(pattern) == 0 || pattern[0] != 1 {
				fmt.Fprintln(os.Stdout, "ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.")
			}
			vec.Vec_IntPush(result, i)
		}
		abc.Abc_NtkDelete(miter)
	}
	if verbose != 0 {
		fmt.Fprintf(os.Stdout, "The outputs are sensitive to %d (out of %d) inputs:\n", vec.Vec_IntSize(result), abc.Abc_NtkCiNum(n))
		for i := 0; i < vec.Vec_IntSize(result); i++ {
			fmt.Fprintf(os.Stdout, "%d ", vec.Vec_IntEntry(result, i))
		}
		fmt.Fprintln(os.Stdout)
	}
	return result
}
