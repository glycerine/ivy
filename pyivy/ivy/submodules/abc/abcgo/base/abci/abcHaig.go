package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkHaigCollectMembers normalizes each marked object's representative and
// returns all non-representative members, matching the source HAIG pass.
func Abc_NtkHaigCollectMembers(manager *Hop_Man_t) *Vec_Ptr_t {
	members := vec.Vec_PtrAlloc(4098)
	for _, object := range hop.Hop_ManObjects(manager) {
		if hop.Hop_ObjData(object) == nil {
			continue
		}
		hop.Hop_ObjSetData(object, hop.Hop_ObjRepr(object))
		vec.Vec_PtrPush(members, object)
	}
	return members
}

func haigDataNode(value any) *Hop_Obj_t {
	node, _ := value.(*Hop_Obj_t)
	return node
}

// Abc_NtkHaigCreateClasses chooses the lowest-ID member as each class head.
func Abc_NtkHaigCreateClasses(members *Vec_Ptr_t) *Vec_Ptr_t {
	classes := vec.Vec_PtrAlloc(4098)
	for _, item := range vec.Vec_PtrArray(members) {
		object := item.(*Hop_Obj_t)
		repr := haigDataNode(hop.Hop_ObjData(object))
		if repr == nil || hop.Hop_ObjData(repr) != nil {
			panic("HAIG representative is not a class root")
		}
		if hop.Hop_ObjIsMarkA(repr) == 0 {
			hop.Hop_ObjSetMarkA(repr)
			vec.Vec_PtrPush(classes, repr)
		}
	}
	for _, item := range vec.Vec_PtrArray(classes) {
		repr := item.(*Hop_Obj_t)
		hop.Hop_ObjClearMarkA(repr)
		hop.Hop_ObjSetData(repr, repr)
	}
	for _, item := range vec.Vec_PtrArray(members) {
		object := item.(*Hop_Obj_t)
		repr := haigDataNode(hop.Hop_ObjData(object))
		current := haigDataNode(hop.Hop_ObjData(repr))
		if current == nil {
			panic("HAIG class lost its initial member")
		}
		if current.Id > object.Id {
			hop.Hop_ObjSetData(repr, object)
		}
	}
	for _, item := range vec.Vec_PtrArray(members) {
		object := item.(*Hop_Obj_t)
		repr := haigDataNode(hop.Hop_ObjData(object))
		chosen := haigDataNode(hop.Hop_ObjData(repr))
		if chosen == nil || chosen.Id > object.Id {
			panic("HAIG representative order is invalid")
		}
		hop.Hop_ObjSetData(object, chosen)
	}
	for i, item := range vec.Vec_PtrArray(classes) {
		old := item.(*Hop_Obj_t)
		repr := haigDataNode(hop.Hop_ObjData(old))
		if repr == nil || haigDataNode(hop.Hop_ObjData(repr)) != repr {
			panic("HAIG class head is not self-representing")
		}
		vec.Vec_PtrWriteEntry(classes, i, repr)
		vec.Vec_PtrPush(members, repr)
	}
	for _, item := range vec.Vec_PtrArray(members) {
		object := item.(*Hop_Obj_t)
		if haigDataNode(hop.Hop_ObjData(object)) == object {
			hop.Hop_ObjSetData(object, nil)
		}
	}
	return classes
}

func Abc_NtkHaigCountFans(manager *Hop_Man_t) int32 {
	var count int32
	for _, object := range hop.Hop_ManObjects(manager) {
		if hop.Hop_ObjData(object) != nil && hop.Hop_ObjRefs(object) > 0 {
			count++
		}
	}
	fmt.Printf("The number of class members with fanouts = %5d.\n", count)
	return count
}

func Hop_ObjReprHop(object *Hop_Obj_t) *Hop_Obj_t {
	next := hop.Hop_ObjReadNext(object)
	if next == nil {
		panic("HAIG HOP node has no reconstructed successor")
	}
	repr := haigDataNode(hop.Hop_ObjData(object))
	if repr == nil {
		return next
	}
	if haigDataNode(hop.Hop_ObjData(repr)) != repr {
		panic("HAIG class representative is not normalized")
	}
	reprNext := hop.Hop_ObjReadNext(repr)
	if reprNext == nil {
		panic("HAIG representative has no reconstructed successor")
	}
	return hop.Hop_NotCond(reprNext, boolInt32(hop.Hop_ObjPhase(object) != hop.Hop_ObjPhase(repr)))
}
func Hop_ObjChild0Hop(object *Hop_Obj_t) *Hop_Obj_t {
	return hop.Hop_NotCond(Hop_ObjReprHop(hop.Hop_ObjFanin0(object)), hop.Hop_ObjFaninC0(object))
}
func Hop_ObjChild1Hop(object *Hop_Obj_t) *Hop_Obj_t {
	return hop.Hop_NotCond(Hop_ObjReprHop(hop.Hop_ObjFanin1(object)), hop.Hop_ObjFaninC1(object))
}

// Abc_NtkHaigReconstruct transfers a HAIG into a fresh HOP manager, replacing
// each class member by its current representative while preserving phase.
func Abc_NtkHaigReconstruct(manager *Hop_Man_t) *Hop_Man_t {
	for _, object := range hop.Hop_ManObjects(manager) {
		hop.Hop_ObjSetNext(object, nil)
	}
	result := hop.Hop_ManStartWithObjects(int(hop.Hop_ManObjNum(manager)) + 1)
	hop.Hop_ObjSetNext(hop.Hop_ManConst1(manager), hop.Hop_ManConst1(result))
	for i := int32(0); i < hop.Hop_ManPiNum(manager); i++ {
		hop.Hop_ObjSetNext(hop.Hop_ManPi(manager, i), hop.Hop_ObjCreatePi(result))
	}
	for _, object := range hop.Hop_ManObjects(manager) {
		if hop.Hop_ObjIsNode(object) == 0 {
			continue
		}
		next := hop.Hop_And(result, Hop_ObjChild0Hop(object), Hop_ObjChild1Hop(object))
		hop.Hop_ObjSetNext(object, next)
		if member := haigDataNode(hop.Hop_ObjData(object)); member != nil {
			reprNext := hop.Hop_ObjReadNext(member)
			if reprNext == nil {
				hop.Hop_ManStop(result)
				return nil
			}
			hop.Hop_ObjSetData(hop.Hop_Regular(next), hop.Hop_Regular(reprNext))
		}
	}
	for i := int32(0); i < hop.Hop_ManPoNum(manager); i++ {
		hop.Hop_ObjCreatePo(result, Hop_ObjChild0Hop(hop.Hop_ManPo(manager, i)))
	}
	if hop.Hop_ManCheck(result) == 0 {
		fmt.Println("Abc_NtkHaigReconstruct: Check for History AIG has failed.")
		hop.Hop_ManStop(result)
		return nil
	}
	return result
}

func Abc_NtkHaigCheckTfi_rec(node, old *Abc_Obj_t, visited map[*Abc_Obj_t]bool) int32 {
	if node == nil {
		return 0
	}
	if node == old {
		return 1
	}
	if abc.Abc_ObjIsCi(node) != 0 {
		return 0
	}
	if abc.Abc_ObjIsNode(node) == 0 {
		panic("HAIG TFI traversal reached a non-node")
	}
	if visited[node] {
		return 0
	}
	visited[node] = true
	for _, fanin := range node.Fanins() {
		if Abc_NtkHaigCheckTfi_rec(fanin.Object, old, visited) != 0 {
			return 1
		}
	}
	for choice := node.Choice; choice != nil; choice = choice.Choice {
		if Abc_NtkHaigCheckTfi_rec(choice, old, visited) != 0 {
			return 1
		}
	}
	return 0
}
func Abc_NtkHaigCheckTfi(network *Abc_Ntk_t, old, node *Abc_Obj_t) int32 {
	if old == nil || node == nil {
		return 0
	}
	return Abc_NtkHaigCheckTfi_rec(node, old, make(map[*Abc_Obj_t]bool))
}

func Hop_ObjChild0NextHaig(object *Hop_Obj_t) abc.Edge {
	fanin := hop.Hop_ObjFanin0(object)
	edge, ok := hop.Hop_ObjReadNextData(fanin).(abc.Edge)
	if !ok {
		panic("HAIG HOP fanin has no ABC successor")
	}
	return abc.Abc_ObjNotCond(edge, hop.Hop_ObjFaninC0(object))
}
func Hop_ObjChild1NextHaig(object *Hop_Obj_t) abc.Edge {
	fanin := hop.Hop_ObjFanin1(object)
	edge, ok := hop.Hop_ObjReadNextData(fanin).(abc.Edge)
	if !ok {
		panic("HAIG HOP fanin has no ABC successor")
	}
	return abc.Abc_ObjNotCond(edge, hop.Hop_ObjFaninC1(object))
}

// Abc_NtkHaigRecreateAig transfers the loop-free HOP graph and its safe choice
// classes back to an ABC strashed network.
func Abc_NtkHaigRecreateAig(network *Abc_Ntk_t, manager *Hop_Man_t) *Abc_Ntk_t {
	result := abc.Abc_NtkStartFrom(network, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	aig, ok := result.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("HAIG output network has no AIG manager")
	}
	hop.Hop_ObjSetNextData(hop.Hop_ManConst1(manager), abc.Abc_AigConst1(result))
	for i := int32(0); i < hop.Hop_ManPiNum(manager); i++ {
		hop.Hop_ObjSetNextData(hop.Hop_ManPi(manager, i), abc.Abc_NtkCi(result, i).Signal())
	}
	for _, object := range hop.Hop_ManObjects(manager) {
		if hop.Hop_ObjIsNode(object) == 0 {
			continue
		}
		edge := abc.Abc_AigAnd(aig, Hop_ObjChild0NextHaig(object), Hop_ObjChild1NextHaig(object))
		if edge.Complement {
			panic("HAIG HOP node reconstructed to a complemented ABC edge")
		}
		hop.Hop_ObjSetNextData(object, edge)
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		old := abc.Abc_NtkCo(network, i)
		po := abc.Abc_ObjCopyEdge(old).Object
		abc.Abc_ObjAddFanin(po, Hop_ObjChild0NextHaig(hop.Hop_ManPo(manager, i)))
	}
	for _, object := range hop.Hop_ManObjects(manager) {
		repr := haigDataNode(hop.Hop_ObjData(object))
		if repr == nil || repr == object || hop.Hop_ObjIsNode(repr) == 0 {
			continue
		}
		thisEdge, okThis := hop.Hop_ObjReadNextData(object).(abc.Edge)
		reprEdge, okRepr := hop.Hop_ObjReadNextData(repr).(abc.Edge)
		if !okThis || !okRepr || thisEdge.Object == nil || reprEdge.Object == nil {
			panic("HAIG choice node has no ABC mapping")
		}
		if thisEdge.Complement || reprEdge.Complement {
			panic("HAIG choice mapping is complemented")
		}
		this, head := thisEdge.Object, reprEdge.Object
		if this.Choice != nil || len(this.Fanouts()) != 0 || head == this {
			panic("HAIG choice candidate violates source invariants")
		}
		if Abc_NtkHaigCheckTfi(result, head, this) != 0 {
			continue
		}
		tail := head
		seen := map[*Abc_Obj_t]bool{head: true}
		for tail.Choice != nil {
			tail = tail.Choice
			if seen[tail] {
				panic("cycle in ABC choice chain")
			}
			seen[tail] = true
		}
		tail.Choice = this
	}
	if abc.Abc_NtkCheck(result) == 0 {
		fmt.Println("Abc_NtkHaigUse: The network check has failed.")
		abc.Abc_NtkDelete(result)
		return nil
	}
	return result
}

func Abc_NtkHaigResetReprsOld(manager *Hop_Man_t) {
	members := Abc_NtkHaigCollectMembers(manager)
	fmt.Printf("Collected %6d class members.\n", vec.Vec_PtrSize(members))
	classes := Abc_NtkHaigCreateClasses(members)
	if vec.Vec_PtrSize(classes) != 0 {
		fmt.Printf("Collected %6d classes. (Ave = %5.2f)\n", vec.Vec_PtrSize(classes), float64(vec.Vec_PtrSize(members))/float64(vec.Vec_PtrSize(classes)))
	} else {
		fmt.Println("Collected      0 classes.")
	}
	vec.Vec_PtrFree(members)
	vec.Vec_PtrFree(classes)
}

// Abc_NtkHaigResetReprs normalizes classes, chooses earliest topological IDs,
// and returns the number of referenced non-representative members.
func Abc_NtkHaigResetReprs(manager *Hop_Man_t) int32 {
	objects := hop.Hop_ManObjects(manager)
	for _, object := range objects {
		repr := haigDataNode(hop.Hop_ObjData(object))
		if repr != nil && haigDataNode(hop.Hop_ObjData(repr)) == object {
			hop.Hop_ObjSetData(repr, repr)
		}
		if repr == object {
			hop.Hop_ObjSetData(object, nil)
		}
	}
	for _, object := range objects {
		if hop.Hop_ObjData(object) == nil {
			continue
		}
		repr := hop.Hop_ObjRepr(object)
		hop.Hop_ObjSetData(repr, repr)
		hop.Hop_ObjSetData(object, repr)
	}
	for _, object := range objects {
		if hop.Hop_ObjData(object) == nil {
			continue
		}
		repr := hop.Hop_ObjRepr(object)
		if repr.Id > object.Id {
			hop.Hop_ObjSetData(repr, object)
			hop.Hop_ObjSetData(object, object)
		} else {
			hop.Hop_ObjSetData(object, repr)
		}
	}
	var fanouts int32
	for _, object := range objects {
		if hop.Hop_ObjData(object) == nil {
			continue
		}
		repr := hop.Hop_ObjRepr(object)
		if hop.Hop_ObjData(object) != repr || repr.Id > object.Id {
			panic("HAIG representative normalization failed")
		}
		if object != repr && hop.Hop_ObjRefs(object) > 0 {
			fanouts++
		}
	}
	return fanouts
}

// Abc_NtkHopRemoveLoops repeatedly collapses classes with live fanout, then
// reconstructs the AIG with only choices that cannot introduce a dependency loop.
func Abc_NtkHopRemoveLoops(network *Abc_Ntk_t, manager *Hop_Man_t) *Abc_Ntk_t {
	for manager != nil && Abc_NtkHaigResetReprs(manager) != 0 {
		temp := manager
		manager = Abc_NtkHaigReconstruct(temp)
		hop.Hop_ManStop(temp)
		if manager == nil {
			return nil
		}
	}
	if manager == nil {
		return nil
	}
	result := Abc_NtkHaigRecreateAig(network, manager)
	hop.Hop_ManStop(manager)
	return result
}
