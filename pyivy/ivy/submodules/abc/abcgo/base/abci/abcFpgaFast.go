package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/aig/ivy"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkFpgaFast maps a strashed network to LUTs and reconstructs the
// mapped cuts as local AIG functions in an ABC logic network.
func Abc_NtkFpgaFast(network *abc.Network, lutSize, recovery, verbose int32) *abc.Network {
	assert(abc.Abc_NtkIsStrash(network) != 0, "Abc_NtkIsStrash(network)")
	manager := Abc_NtkIvyBefore(network, 0, 0)
	if manager == nil {
		return nil
	}
	ivy.Ivy_FastMapPerform(manager, lutSize, recovery, verbose)
	mapped := Ivy_ManFpgaToAbc(network, manager)
	ivy.Ivy_FastMapStop(manager)
	ivy.Ivy_ManStop(manager)
	if mapped != nil && abc.Abc_NtkCheck(mapped) == 0 {
		fmt.Println("Abc_NtkFastMap: The network check has failed.")
		abc.Abc_NtkDelete(mapped)
		return nil
	}
	return mapped
}

// Ivy_ManFpgaToAbc reconstructs the mapper's LUT cuts in ABC's Hop functions.
func Ivy_ManFpgaToAbc(source *abc.Network, manager *ivy.Ivy_Man_t) *abc.Network {
	ivy.Ivy_ManSetCopy(manager, vec.Vec_PtrStart(int(ivy.Ivy_ManObjIdMax(manager)+1)))
	mapped := abc.Abc_NtkStartFrom(source, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG)
	Abc_ObjSetIvy2Abc(manager, ivy.Ivy_ManConst1(manager).Id, abc.Abc_NtkCreateNodeConst1(mapped))
	for i := int32(0); i < abc.Abc_NtkCiNum(mapped); i++ {
		Abc_ObjSetIvy2Abc(manager, ivy.Ivy_ManPi(manager, i).Id, abc.Abc_NtkCi(mapped, i))
	}
	cutNodes := vec.Vec_IntAlloc(100)
	for i := int32(0); i < ivy.Ivy_ManPoNum(manager); i++ {
		po := ivy.Ivy_ManPo(manager, i)
		driver := Ivy_ManToAbcFast_rec(mapped, manager, ivy.Ivy_ObjFanin0(po), cutNodes)
		if ivy.Ivy_ObjFaninC0(po) != 0 {
			if abc.Abc_ObjIsCi(driver) != 0 {
				driver = abc.Abc_NtkCreateNodeInv(mapped, driver)
			} else {
				inverted := abc.Abc_NtkCloneObj(driver)
				function, ok := driver.Data.(*hop.Hop_Obj_t)
				assert(ok, "mapped node has a Hop function")
				inverted.Data = hop.Hop_Not(function)
				driver = inverted
			}
		}
		abc.Abc_ObjAddFanin(abc.Abc_NtkCo(mapped, i), driver)
	}
	vec.Vec_IntFree(cutNodes)
	vec.Vec_PtrFree(ivy.Ivy_ManReadCopy(manager).(*vec.Vec_Ptr_t))
	ivy.Ivy_ManSetCopy(manager, nil)
	abc.Abc_NtkCleanup(mapped, 0)
	abc.Abc_NtkLogicMakeSimpleCos(mapped, 0)
	return mapped
}

// Ivy_ManToAbcFast_rec recursively reconstructs one LUT root. The scratch
// vector is the source-compatible cut worklist reused by all output cones.
func Ivy_ManToAbcFast_rec(network *abc.Network, manager *ivy.Ivy_Man_t, root *ivy.Ivy_Obj_t, cutNodes *vec.Vec_Int_t) *abc.Object {
	if mapped := Abc_ObjGetIvy2Abc(manager, root.Id); mapped != nil {
		return mapped
	}
	assert(ivy.Ivy_ObjIsAnd(root) != 0 || ivy.Ivy_ObjIsExor(root) != 0, "root is an Ivy AND or XOR")
	var support []int32
	ivy.Ivy_FastMapReadSupp(manager, root, &support)
	leaves := vec.Vec_IntAllocArrayCopy(support, len(support))
	node := abc.Abc_NtkCreateNode(network)
	functions := make(map[int32]*hop.Hop_Obj_t, len(support))
	for index, leafID := range support {
		fanin := Ivy_ManToAbcFast_rec(network, manager, ivy.Ivy_ManObj(manager, leafID), cutNodes)
		abc.Abc_ObjAddFanin(node, fanin)
		functions[leafID] = hop.Hop_IthVar(network.FunctionManager.(*hop.Hop_Man_t), int32(index))
	}
	ivy.Ivy_ManCollectCut(manager, root, leaves, cutNodes)
	for i := int32(len(support)); i < int32(vec.Vec_IntSize(cutNodes)); i++ {
		ivyNode := ivy.Ivy_ManObj(manager, vec.Vec_IntEntry(cutNodes, int(i)))
		left := functions[ivy.Ivy_ObjFanin0(ivyNode).Id]
		right := functions[ivy.Ivy_ObjFanin1(ivyNode).Id]
		assert(left != nil && right != nil, "cut fanins have local functions")
		left = hop.Hop_NotCond(left, ivy.Ivy_ObjFaninC0(ivyNode))
		right = hop.Hop_NotCond(right, ivy.Ivy_ObjFaninC1(ivyNode))
		function := hop.Hop_And(network.FunctionManager.(*hop.Hop_Man_t), left, right)
		functions[ivyNode.Id] = function
	}
	function := functions[root.Id]
	assert(function != nil, "root local function was derived")
	node.Data = function
	Abc_ObjSetIvy2Abc(manager, root.Id, node)
	vec.Vec_IntFree(leaves)
	return node
}

func Abc_ObjSetIvy2Abc(manager *ivy.Ivy_Man_t, ivyID int32, object *abc.Object) {
	copy := ivy.Ivy_ManReadCopy(manager).(*vec.Vec_Ptr_t)
	assert(vec.Vec_PtrEntry(copy, int(ivyID)) == nil, "Ivy object has no prior ABC copy")
	assert(abc.Abc_ObjIsComplement(object) == 0, "ABC copy is regular")
	vec.Vec_PtrWriteEntry(copy, int(ivyID), object)
}

func Abc_ObjGetIvy2Abc(manager *ivy.Ivy_Man_t, ivyID int32) *abc.Object {
	object, _ := vec.Vec_PtrEntry(ivy.Ivy_ManReadCopy(manager).(*vec.Vec_Ptr_t), int(ivyID)).(*abc.Object)
	return object
}
