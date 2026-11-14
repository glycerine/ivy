package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
)

const attachFull uint32 = ^uint32(0)

// Abc_NtkAttach replaces each SOP node by a matching gate from the active
// genlib. Gate matching and fanin permutation follow abcAttach.c.
func Abc_NtkAttach(n *Abc_Ntk_t) int32 {
	if abc.Abc_NtkIsSopLogic(n) == 0 {
		panic("Abc_NtkAttach requires SOP logic")
	}
	lib, _ := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
	if lib == nil {
		fmt.Println("The current library is not available.")
		return 0
	}
	var truths [6][2]uint32
	Abc_AttachSetupTruthTables(truths[:])
	count := int32(0)
	gates := mio.Mio_CollectRoots(lib, 6, 1.0e20, 1, &count, 0)
	gateTruths := make([][2]uint32, count)
	for i, g := range gates {
		mio.Mio_DeriveTruthTable(g, truths[:], mio.Mio_GateReadPinNum(g), 6, gateTruths[i][:])
	}
	abc.Abc_NtkCleanCopy(n)
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		sop, _ := node.Data.(string)
		fanins := abc.Abc_ObjFaninNum(node)
		switch {
		case fanins == 0:
			if abc.Abc_SopIsConst1(sop) != 0 {
				abc.Abc_ObjSetCopyData(node, libConst1(lib))
			} else {
				abc.Abc_ObjSetCopyData(node, libConst0(lib))
			}
		case fanins == 1:
			if abc.Abc_SopIsBuf(sop) != 0 {
				abc.Abc_ObjSetCopyData(node, libBuf(lib))
			} else {
				abc.Abc_ObjSetCopyData(node, libInv(lib))
			}
		case fanins > 6:
			fmt.Printf("Cannot attach gate with more than 6 inputs to node %s.\n", abc.Abc_ObjName(node))
			return 0
		default:
			if Abc_NodeAttach(n, node, gates, gateTruths, truths[:]) == 0 {
				fmt.Printf("Could not attach the library gate to node %s.\n", abc.Abc_ObjName(node))
				return 0
			}
		}
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if abc.Abc_ObjCopyData(node) == nil {
			fmt.Println("Some elementary gates (constant, buffer, or inverter) are missing in the library.")
			return 0
		}
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		node.Data = abc.Abc_ObjCopyData(node)
		abc.Abc_ObjSetCopyData(node, nil)
	}
	n.Function = uint8(abc.ABC_FUNC_MAP)
	if manager, ok := n.FunctionManager.(*mem.Mem_Flex_t); ok {
		mem.Mem_FlexStop(manager, 0)
	}
	n.FunctionManager = lib
	if !abc.Abc_NtkCheck(n) {
		fmt.Println("Abc_NtkAttach: The network check has failed.")
		return 0
	}
	fmt.Println("Library gates are successfully attached to the nodes.")
	return 1
}

// Abc_AttachTransferFanins applies the gate input permutation while preserving
// edge polarity and keeping the network's fanout indexes synchronized.
func Abc_AttachTransferFanins(n *Abc_Ntk_t, node *Abc_Obj_t, fanins []abc.Edge, perm []int32, count int32) {
	if count < 0 || int(count) > len(fanins) || int(count) > len(perm) {
		panic("invalid attachment fanin count")
	}
	for i := int32(0); i < count; i++ {
		if perm[i] < 0 || perm[i] >= count {
			panic("invalid attachment permutation")
		}
		if err := n.PatchFanin(node, int(perm[i]), fanins[i]); err != nil {
			panic(err)
		}
	}
}

func libConst0(l *mio.Mio_Library_t) *mio.Mio_Gate_t { return mio.Mio_LibraryReadConst0(l) }
func libConst1(l *mio.Mio_Library_t) *mio.Mio_Gate_t { return mio.Mio_LibraryReadConst1(l) }
func libBuf(l *mio.Mio_Library_t) *mio.Mio_Gate_t    { return mio.Mio_LibraryReadBuf(l) }
func libInv(l *mio.Mio_Library_t) *mio.Mio_Gate_t    { return mio.Mio_LibraryReadInv(l) }

func Abc_AttachSetupTruthTables(t [][2]uint32) {
	for v := 0; v < 5; v++ {
		t[v][0] = 0
	}
	for m := 0; m < 32; m++ {
		for v := 0; v < 5; v++ {
			if m&(1<<v) != 0 {
				t[v][0] |= uint32(1) << uint(m)
			}
		}
	}
	for v := 0; v < 5; v++ {
		t[v][1] = t[v][0]
	}
	t[5][0] = 0
	t[5][1] = attachFull
}

func Abc_AttachComputeTruth(sop string, inputs [][2]uint32, result []uint32) {
	result[0], result[1] = 0, 0
	nvars := int(abc.Abc_SopGetVarNum(sop))
	if nvars < 0 {
		panic("malformed SOP")
	}
	for pos := 0; pos < len(sop); {
		end := pos
		for end < len(sop) && sop[end] != '\n' {
			end++
		}
		if end == pos {
			break
		}
		var c [2]uint32
		c[0], c[1] = attachFull, attachFull
		for k := 0; k < nvars; k++ {
			switch sop[pos+k] {
			case '0':
				c[0] &= ^inputs[k][0]
				c[1] &= ^inputs[k][1]
			case '1':
				c[0] &= inputs[k][0]
				c[1] &= inputs[k][1]
			}
		}
		result[0] |= c[0]
		result[1] |= c[1]
		pos = end + 1
	}
	if abc.Abc_SopGetPhase(sop) == 0 {
		result[0] = ^result[0]
		result[1] = ^result[1]
	}
}

func Abc_AttachFind(gates []*mio.Mio_Gate_t, truths [][2]uint32, node []uint32, perm []int32) *mio.Mio_Gate_t {
	if i := Abc_AttachCompare(truths, node); i >= 0 {
		for v := 0; v < 6; v++ {
			perm[v] = int32(v)
		}
		return gates[i]
	}
	for _, p := range extra.Extra_Permutations(6) {
		var tt [2]uint32
		Abc_TruthPermute(p, 6, node, tt[:])
		if i := Abc_AttachCompare(truths, tt[:]); i >= 0 {
			for v := 0; v < 6; v++ {
				perm[v] = int32(p[v])
			}
			return gates[i]
		}
	}
	return nil
}

func Abc_AttachCompare(truths [][2]uint32, node []uint32) int32 {
	for i, truth := range truths {
		if truth[0] == node[0] && truth[1] == node[1] {
			return int32(i)
		}
	}
	return -1
}

func Abc_NodeAttach(n *Abc_Ntk_t, node *Abc_Obj_t, gates []*mio.Mio_Gate_t, truths [][2]uint32, inputs [][2]uint32) int32 {
	sop, _ := node.Data.(string)
	var table [2]uint32
	Abc_AttachComputeTruth(sop, inputs, table[:])
	perm := make([]int32, 6)
	gate := Abc_AttachFind(gates, truths, table[:], perm)
	if gate == nil {
		return 0
	}
	count := abc.Abc_ObjFaninNum(node)
	fanins := make([]abc.Edge, count)
	for i := int32(0); i < count; i++ {
		fanins[i] = abc.Abc_ObjChild(node, i)
	}
	Abc_AttachTransferFanins(n, node, fanins, perm, count)
	abc.Abc_ObjSetCopyData(node, gate)
	return 1
}

func Abc_TruthPermute(p []byte, nvars int32, node, result []uint32) {
	result[0], result[1] = 0, 0
	for mint := 0; mint < (1 << uint(nvars)); mint++ {
		if node[mint>>5]&(uint32(1)<<uint(mint&31)) == 0 {
			continue
		}
		out := 0
		for v := int32(0); v < nvars; v++ {
			if mint&(1<<uint(v)) != 0 {
				out |= 1 << uint(p[v])
			}
		}
		result[out>>5] |= uint32(1) << uint(out&31)
	}
}
