package abci

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/st"
)

type dressMapping = *st.Table

// Abc_NtkDress transfers names from a reference netlist onto the current
// logic network by strashing and fraiging their combined networks.
func Abc_NtkDress(logic *Abc_Ntk_t, fileName string, verbose int32) {
	assert(baseabc.Abc_NtkIsLogic(logic) != 0, "Abc_NtkIsLogic(logic)")

	original := abcio.Io_ReadNetlist(fileName, abcio.Io_ReadFileType(fileName), 1)
	if original == nil {
		return
	}
	assert(baseabc.Abc_NtkIsNetlist(original) != 0, "Abc_NtkIsNetlist(original)")

	baseabc.Abc_NtkCleanCopy(logic)
	baseabc.Abc_NtkCleanCopy(original)
	logicOriginal := baseabc.Abc_NtkToLogic(original)
	if logicOriginal == nil {
		baseabc.Abc_NtkDelete(original)
		return
	}
	if baseabc.Abc_NtkCompareSignals(logic, logicOriginal, 1, 1) == 0 {
		baseabc.Abc_NtkDelete(original)
		baseabc.Abc_NtkDelete(logicOriginal)
		return
	}

	miter := Abc_NtkStrash(logic, 1, 0, 0)
	if miter == nil {
		baseabc.Abc_NtkDelete(original)
		baseabc.Abc_NtkDelete(logicOriginal)
		return
	}
	Abc_NtkAppend(miter, logicOriginal, 1)
	baseabc.Abc_NtkTransferCopy(original)
	baseabc.Abc_NtkDelete(logicOriginal)
	if verbose != 0 {
		fmt.Printf("After mitering:\n")
		fmt.Printf("Logic:  Nodes = %5d. Copy = %5d. \n", baseabc.Abc_NtkNodeNum(logic), baseabc.Abc_NtkCountCopy(logic))
		fmt.Printf("Orig:   Nodes = %5d. Copy = %5d. \n", baseabc.Abc_NtkNodeNum(original), baseabc.Abc_NtkCountCopy(original))
	}

	miterFraig := Abc_NtkIvyFraig(miter, 100, 1, 0, 1, 0)
	baseabc.Abc_NtkTransferCopy(logic)
	baseabc.Abc_NtkTransferCopy(original)
	baseabc.Abc_NtkDelete(miter)
	if miterFraig == nil {
		baseabc.Abc_NtkDelete(original)
		return
	}
	if verbose != 0 {
		fmt.Printf("After fraiging:\n")
		fmt.Printf("Logic:  Nodes = %5d. Copy = %5d. \n", baseabc.Abc_NtkNodeNum(logic), baseabc.Abc_NtkCountCopy(logic))
		fmt.Printf("Orig:   Nodes = %5d. Copy = %5d. \n", baseabc.Abc_NtkNodeNum(original), baseabc.Abc_NtkCountCopy(original))
	}

	mapping := Abc_NtkDressDeriveMapping(original)
	Abc_NtkDressTransferNames(logic, mapping, verbose)
	baseabc.Abc_NtkDelete(miterFraig)
	baseabc.Abc_NtkDelete(original)
}

// Abc_NtkDressDeriveMapping maps each regular fraig node to its first
// prototype node in the original netlist, retaining the source polarity.
func Abc_NtkDressDeriveMapping(network *Abc_Ntk_t) dressMapping {
	assert(baseabc.Abc_NtkIsNetlist(network) != 0, "Abc_NtkIsNetlist(network)")
	result := st.New(st.PointerCompare, st.PointerHash)
	for _, node := range network.ObjectsOfKind(baseabc.ObjectNode) {
		copy := baseabc.Abc_ObjCopyEdge(node)
		fraig := baseabc.Abc_ObjRegular(copy)
		assert(fraig != nil, "original node has no fraig copy")
		if _, exists := result.Lookup(fraig); exists {
			continue
		}
		result.Insert(fraig, baseabc.Abc_ObjNotCond(node, baseabc.Abc_ObjIsComplement(copy)))
	}
	return result
}

// Abc_NtkDressTransferNames assigns each available prototype net name once.
// Existing logic-node names win; among duplicate fraig prototypes, derivation
// order chooses the first source node exactly as the C hash insertion does.
func Abc_NtkDressTransferNames(network *Abc_Ntk_t, mapping dressMapping, verbose int32) {
	assert(baseabc.Abc_NtkIsLogic(network) != 0, "Abc_NtkIsLogic(network)")
	initial := mapping.Count()
	assigned, complemented := 0, 0
	for _, node := range network.ObjectsOfKind(baseabc.ObjectNode) {
		if node.Name() != "" {
			continue
		}
		copy := baseabc.Abc_ObjCopyEdge(node)
		fraig := baseabc.Abc_ObjRegular(copy)
		assert(fraig != nil, "logic node has no fraig copy")
		value, exists := mapping.Lookup(fraig)
		if !exists {
			continue
		}
		prototype, ok := value.(baseabc.Edge)
		assert(ok, "prototype mapping has unexpected type")
		prototype = baseabc.Abc_ObjNotCond(prototype, baseabc.Abc_ObjIsComplement(copy))
		net := baseabc.Abc_ObjFanout0(baseabc.Abc_ObjRegular(prototype))
		assert(net != nil, "prototype node has no fanout net")
		name := net.Name()
		assert(name != "", "source prototype net has no name")
		suffix := ""
		if baseabc.Abc_ObjIsComplement(prototype) != 0 {
			suffix = "_inv"
			complemented++
		} else {
			assigned++
		}
		baseabc.Abc_ObjAssignName(node, name, suffix)
		mapping.Delete(fraig)
	}
	if verbose != 0 {
		fmt.Printf("Total number of names collected = %5d.\n", initial)
		fmt.Printf("Total number of names assigned  = %5d. (Dir = %5d. Compl = %5d.)\n", assigned+complemented, assigned, complemented)
	}
}
