package wln

// Network utilities ported from src/base/wln/wlnNtk.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"unsafe"
)

// cWlnVec/cWlnNtk describe the C ABI for the source's memory estimate. They
// are not the Go network representation.
type cWlnVec struct {
	size, cap int32
	data      uintptr
}
type cWlnNtk struct {
	name, spec                                             uintptr
	smt                                                    int32
	cis, cos, ffs, types                                   cWlnVec
	fanins                                                 uintptr
	ranges                                                 cWlnVec
	rangeTable                                             uintptr
	names, inst                                            cWlnVec
	nameManager                                            uintptr
	signs                                                  cWlnVec
	travID                                                 int32
	trav, copies, bits, levels, refs, fanout, attrs, lists cWlnVec
	tables                                                 uintptr
	objs, ands                                             [ABC_OPER_LAST]int32
}

func Wln_NtkMemUsage(p *Wln_Ntk_t) int32 {
	n := int32(unsafe.Sizeof(cWlnNtk{}))
	for _, v := range []*Vec_Int_t{&p.vCis, &p.vCos, &p.vFfs, &p.vRanges, &p.vNameIds, &p.vInstIds, &p.vTravIds, &p.vCopies, &p.vBits, &p.vLevels, &p.vRefs, &p.vFanout, &p.vFaninAttrs, &p.vFaninLists} {
		n += 4 * Vec_IntCap(v)
	}
	n += Vec_IntCap(&p.vTypes) + 20*int32(vec.Hash_IntManEntryNum(p.pRanges)) + util.Abc_NamMemUsed(p.pManName)
	return n
}
func Wln_NtkPrint(p *Wln_Ntk_t) {
	fmt.Printf("Printing %d objects of network \"%s\":\n", Wln_NtkObjNum(p), p.pName)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		Wln_ObjPrint(p, id)
	}
	fmt.Println()
}
func Wln_NtkIsAcyclic_rec(p *Wln_Ntk_t, id int32) int32 {
	if Wln_ObjIsTravIdPrevious(p, id) != 0 {
		return 1
	}
	if Wln_ObjIsTravIdCurrent(p, id) != 0 {
		fmt.Println("Network contains combinational loop!")
		fmt.Printf("Node %16s is encountered twice on the following path:\n", Wln_ObjName(p, id))
		wlnPrintCycleNode(p, id)
		return 0
	}
	Wln_ObjSetTravIdCurrent(p, id)
	if Wln_ObjIsCi(p, id) != 0 || Wln_ObjIsFf(p, id) != 0 || Wln_ObjFaninNum(p, id) == 0 {
		Wln_ObjSetTravIdPrevious(p, id)
		return 1
	}
	for _, f := range Wln_ObjFanins(p, id)[:Wln_ObjFaninNum(p, id)] {
		if f != 0 && Wln_NtkIsAcyclic_rec(p, f) == 0 {
			wlnPrintCycleNode(p, id)
			return 0
		}
	}
	Wln_ObjSetTravIdPrevious(p, id)
	return 1
}
func wlnPrintCycleNode(p *Wln_Ntk_t, id int32) {
	fmt.Printf("Node %16s (ID %6d) of type %5s (type ID %2d) ->\n", Wln_ObjName(p, id), id, miniaig.Abc_OperName(int(Wln_ObjType(p, id))), Wln_ObjType(p, id))
}
func Wln_NtkIsAcyclic(p *Wln_Ntk_t) int32 {
	Wln_NtkIncrementTravId(p)
	Wln_NtkIncrementTravId(p)
	for i := int32(0); i < Wln_NtkCoNum(p); i++ {
		id := Wln_NtkCo(p, i)
		if Wln_NtkIsAcyclic_rec(p, id) == 0 {
			fmt.Printf("Primary output %16s (ID %6d)\n", Wln_ObjName(p, id), id)
			return 0
		}
	}
	for i := int32(0); i < Wln_NtkFfNum(p); i++ {
		id := Wln_NtkFf(p, i)
		if Wln_NtkIsAcyclic_rec(p, id) == 0 {
			fmt.Printf("Flip-flop %16s (ID %6d)\n", Wln_ObjName(p, id), id)
			return 0
		}
	}
	var unvisited int32
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		if Wln_ObjIsTravIdPrevious(p, id) == 0 && Wln_ObjIsCi(p, id) == 0 {
			unvisited++
		}
	}
	if unvisited != 0 {
		Wln_NtkCreateRefs(p)
		fmt.Printf("The network has %d objects and %d (%6.2f %%) of them are not connected to the outputs.\n", Wln_NtkObjNum(p), unvisited, 100*float64(unvisited)/float64(Wln_NtkObjNum(p)))
		var sinks []int32
		for id := int32(1); id < Wln_NtkObjNum(p); id++ {
			if Wln_ObjRefs(p, id) == 0 && Wln_ObjIsCi(p, id) == 0 && Wln_ObjIsCo(p, id) == 0 && Wln_ObjIsFf(p, id) == 0 {
				sinks = append(sinks, id)
			}
		}
		if len(sinks) > 0 {
			fmt.Printf("These unconnected objects feed into %d sink objects without fanout:\n", len(sinks))
			for i, id := range sinks {
				if i == 5 {
					break
				}
				fmt.Printf("Node %16s (ID %6d) of type %5s (type ID %2d)\n", Wln_ObjName(p, id), id, miniaig.Abc_OperName(int(Wln_ObjType(p, id))), Wln_ObjType(p, id))
			}
			if len(sinks) > 5 {
				fmt.Println("...")
			}
		}
		for id := int32(1); id < Wln_NtkObjNum(p); id++ {
			if Wln_ObjIsTravIdPrevious(p, id) == 0 && Wln_ObjIsCi(p, id) == 0 && Wln_NtkIsAcyclic_rec(p, id) == 0 {
				fmt.Printf("Unconnected object %s\n", Wln_ObjName(p, id))
				return 0
			}
		}
	}
	return 1
}
func Wln_NtkTransferNames(dst, src *Wln_Ntk_t) {
	assert(dst.pManName == nil && src.pManName != nil, "name manager ownership")
	dst.pManName = src.pManName
	src.pManName = nil
	assert(Wln_NtkHasCopy(dst) == 0 && Wln_NtkHasCopy(src) != 0, "copy map ownership")
	if Wln_NtkHasNameId(src) != 0 {
		assert(Wln_NtkHasNameId(dst) == 0, "destination names")
		Wln_NtkCleanNameId(dst)
		for id := int32(1); id < Wln_NtkObjNum(src); id++ {
			if Wln_ObjCopy(src, id) != 0 && id < Vec_IntSize(&src.vNameIds) && Wln_ObjNameId(src, id) != 0 {
				Wln_ObjSetNameId(dst, Wln_ObjCopy(src, id), Wln_ObjNameId(src, id))
			}
		}
		vec.Vec_IntErase(&src.vNameIds)
	}
	if Wln_NtkHasInstId(src) != 0 {
		assert(Wln_NtkHasInstId(dst) == 0, "destination instances")
		Wln_NtkCleanInstId(dst)
		for id := int32(1); id < Wln_NtkObjNum(src); id++ {
			if Wln_ObjCopy(src, id) != 0 && id < Vec_IntSize(&src.vInstIds) && Wln_ObjInstId(src, id) != 0 {
				Wln_ObjSetInstId(dst, Wln_ObjCopy(src, id), Wln_ObjInstId(src, id))
			}
		}
		vec.Vec_IntErase(&src.vInstIds)
	}
}
func Wln_ObjDup(dst, src *Wln_Ntk_t, id int32) int32 {
	result := Wln_ObjClone(dst, src, id)
	for _, f := range Wln_ObjFanins(src, id)[:Wln_ObjFaninNum(src, id)] {
		if f != 0 {
			Wln_ObjAddFanin(dst, result, Wln_ObjCopy(src, f))
		}
	}
	if Wln_ObjIsConst(src, id) != 0 {
		Wln_ObjSetConst(dst, result, Wln_ObjFanin0(src, id))
	} else if Wln_ObjIsSlice(src, id) != 0 || Wln_ObjIsRotate(src, id) != 0 || Wln_ObjIsTable(src, id) != 0 {
		Wln_ObjSetFanin(dst, result, 1, Wln_ObjFanin1(src, id))
	}
	Wln_ObjSetCopy(src, id, result)
	return result
}
func Wln_NtkDupDfs_rec(dst, src *Wln_Ntk_t, id int32) int32 {
	if id == 0 {
		return 0
	}
	if Wln_ObjCopy(src, id) != 0 {
		return Wln_ObjCopy(src, id)
	}
	assert(Wln_ObjIsFf(src, id) == 0, "FF in combinational cone")
	for _, f := range Wln_ObjFanins(src, id)[:Wln_ObjFaninNum(src, id)] {
		if f != 0 {
			Wln_NtkDupDfs_rec(dst, src, f)
		}
	}
	return Wln_ObjDup(dst, src, id)
}
func Wln_NtkDupDfs(src *Wln_Ntk_t) *Wln_Ntk_t {
	dst := Wln_NtkAlloc(src.pName, Wln_NtkObjNum(src))
	dst.fSmtLib = src.fSmtLib
	dst.pSpec = src.pSpec
	Wln_NtkCleanCopy(src)
	for i := int32(0); i < Wln_NtkCiNum(src); i++ {
		Wln_ObjDup(dst, src, Wln_NtkCi(src, i))
	}
	for i := int32(0); i < Wln_NtkFfNum(src); i++ {
		id := Wln_NtkFf(src, i)
		Wln_ObjSetCopy(src, id, Wln_ObjClone(dst, src, id))
	}
	for i := int32(0); i < Wln_NtkCoNum(src); i++ {
		Wln_NtkDupDfs_rec(dst, src, Wln_NtkCo(src, i))
	}
	for i := int32(0); i < Wln_NtkFfNum(src); i++ {
		id := Wln_NtkFf(src, i)
		for _, f := range Wln_ObjFanins(src, id)[:Wln_ObjFaninNum(src, id)] {
			if f != 0 {
				Wln_ObjAddFanin(dst, Wln_ObjCopy(src, id), Wln_NtkDupDfs_rec(dst, src, f))
			}
		}
	}
	if Wln_NtkHasNameId(src) != 0 {
		Wln_NtkTransferNames(dst, src)
	}
	return dst
}
func Wln_NtkCreateRefs(p *Wln_Ntk_t) {
	Wln_NtkCleanRefs(p)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		for _, f := range Wln_ObjFanins(p, id)[:Wln_ObjFaninNum(p, id)] {
			if f != 0 {
				Wln_ObjRefsInc(p, f)
			}
		}
	}
}
func Wln_NtkFaninNum(p *Wln_Ntk_t) int32 {
	var n int32
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		n += Wln_ObjFaninNum(p, id)
	}
	return n
}
func Wln_NtkStartFaninMap(p *Wln_Ntk_t, m *Vec_Int_t, multi int32) {
	offset := Wln_NtkObjNum(p)
	Vec_IntFill(m, offset+multi*Wln_NtkFaninNum(p), 0)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		Vec_IntWriteEntry(m, id, offset)
		offset += multi * Wln_ObjFaninNum(p, id)
	}
	assert(offset == Vec_IntSize(m), "fanin map size")
}
func Wln_NtkStartFanoutMap(p *Wln_Ntk_t, m, counts *Vec_Int_t, multi int32) {
	offset := Wln_NtkObjNum(p)
	Vec_IntFill(m, offset+multi*vec.Vec_IntSum(counts), 0)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		Vec_IntWriteEntry(m, id, offset)
		offset += multi * Wln_ObjRefs(p, id)
	}
	assert(offset == Vec_IntSize(m), "fanout map size")
}
func Wln_NtkStaticFanoutStart(p *Wln_Ntk_t) {
	Wln_NtkCreateRefs(p)
	Wln_NtkStartFanoutMap(p, &p.vFanout, &p.vRefs, 1)
	old := append([]int32(nil), vec.Vec_IntArray(&p.vRefs)...)
	Wln_NtkCleanRefs(p)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		for _, f := range Wln_ObjFanins(p, id)[:Wln_ObjFaninNum(p, id)] {
			if f != 0 {
				Wln_ObjSetFanout(p, f, Wln_ObjRefsInc(p, f), id)
			}
		}
	}
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		assert(Wln_ObjRefs(p, id) == old[id], "fanout count")
	}
}
func Wln_NtkStaticFanoutStop(p *Wln_Ntk_t) { vec.Vec_IntErase(&p.vRefs); vec.Vec_IntErase(&p.vFanout) }
func Wln_NtkStaticFanoutTest(p *Wln_Ntk_t) {
	fmt.Printf("Printing fanouts of %d objects of network \"%s\":\n", Wln_NtkObjNum(p), p.pName)
	Wln_NtkStaticFanoutStart(p)
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		Wln_ObjPrint(p, id)
		fmt.Print("   Fanouts : ")
		for i := int32(0); i < Wln_ObjFanoutNum(p, id); i++ {
			fmt.Printf("%5d ", Wln_ObjFanout(p, id, i))
		}
		fmt.Println()
	}
	Wln_NtkStaticFanoutStop(p)
	fmt.Println()
}
