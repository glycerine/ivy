package gia

import (
	"fmt"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

var Iso_Nodes = [6]uint32{0x04892ed6, 0xc2523d7d, 0xdc36cd2e, 0xf2db64f8, 0xde3126bb, 0xdebbdff0}
var Iso_Fanio = [2]uint32{0x855ee0cf, 0x946e1b5f}
var Iso_Compl = [2]uint32{0x8ba63e50, 0x14d87f02}

func Gia_Iso3Node(obj *Gia_Obj_t) uint32 {
	if Gia_ObjIsAnd(obj) != 0 {
		return Iso_Nodes[Gia_ObjFaninC0(obj)+Gia_ObjFaninC1(obj)]
	}
	if Gia_ObjIsCi(obj) != 0 {
		return Iso_Nodes[3]
	}
	if Gia_ObjIsCo(obj) != 0 {
		return Iso_Nodes[4]
	}
	return Iso_Nodes[5]
}
func Gia_Iso3Init(g *Gia_Man_t) {
	for _, obj := range g.pObjs {
		obj.Value = Gia_Iso3Node(obj)
	}
}
func Gia_Iso3ComputeEdge(g *Gia_Man_t, obj, fanin *Gia_Obj_t, complement int32, signatures *Vec_Int_t) {
	if complement < 0 || complement > 1 {
		panic("gia: invalid edge phase")
	}
	obj.Value += uint32(vec.Vec_IntEntry(signatures, int(Gia_ObjId(g, fanin)))) + Iso_Compl[complement] + Iso_Fanio[0]
	fanin.Value += uint32(vec.Vec_IntEntry(signatures, int(Gia_ObjId(g, obj)))) + Iso_Compl[complement] + Iso_Fanio[1]
}
func Gia_Iso3Compute(g *Gia_Man_t, signatures *Vec_Int_t) {
	if vec.Vec_IntSize(signatures) != int(Gia_ManObjNum(g)) {
		panic("gia: signature count mismatch")
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) != 0 || Gia_ObjIsCo(obj) != 0 {
			Gia_Iso3ComputeEdge(g, obj, Gia_ObjFanin0(obj), Gia_ObjFaninC0(obj), signatures)
		}
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_Iso3ComputeEdge(g, obj, Gia_ObjFanin1(obj), Gia_ObjFaninC1(obj), signatures)
		}
	}
}
func Gia_Iso3Save(g *Gia_Man_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(Gia_ManObjNum(g)))
	for _, obj := range g.pObjs {
		vec.Vec_IntPush(result, int32(obj.Value))
	}
	return result
}
func Gia_Iso3Unique(signatures *Vec_Int_t) int32 {
	seen := make(map[int32]struct{}, vec.Vec_IntSize(signatures))
	for _, value := range vec.Vec_IntArray(signatures) {
		seen[value] = struct{}{}
	}
	return int32(len(seen))
}
func Gia_Iso3Test(g *Gia_Man_t) {
	const maxIter = 500
	start := time.Now()
	previous := int32(-1)
	Gia_Iso3Init(g)
	for i := 0; i < maxIter; i++ {
		signatures := Gia_Iso3Save(g)
		count := Gia_Iso3Unique(signatures)
		fmt.Printf("Iter %3d : %6d  out of %6d  Time = %.2f sec\n", i, count, vec.Vec_IntSize(signatures), time.Since(start).Seconds())
		if count == previous {
			break
		}
		previous = count
		Gia_Iso3Compute(g, signatures)
	}
}

// Gia_Iso4Gia assigns reverse-level randomized structural signatures.
func Gia_Iso4Gia(g *Gia_Man_t) *Vec_Wec_t {
	levels := Gia_ManLevelizeR(g)
	abcutil.Abc_Random(1)
	for level := 0; level < vec.Vec_WecSize(levels); level++ {
		constants := [2]uint32{abcutil.Abc_Random(0), abcutil.Abc_Random(0)}
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(levels, level)) {
			obj := Gia_ManObj(g, id)
			if level == 0 {
				if Gia_ObjIsCo(obj) == 0 {
					panic("gia: reverse level zero is not an output")
				}
				obj.Value = abcutil.Abc_Random(0)
				Gia_ObjFanin0(obj).Value += obj.Value + constants[Gia_ObjFaninC0(obj)]
			} else if Gia_ObjIsAnd(obj) != 0 {
				Gia_ObjFanin0(obj).Value += obj.Value + constants[Gia_ObjFaninC0(obj)]
				Gia_ObjFanin1(obj).Value += obj.Value + constants[Gia_ObjFaninC1(obj)]
			}
		}
	}
	return levels
}
func Gia_Iso4Test(g *Gia_Man_t) {
	levels := Gia_Iso4Gia(g)
	for level := 0; level < vec.Vec_WecSize(levels); level++ {
		fmt.Printf("Level %d\n", level)
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(levels, level)) {
			fmt.Printf("Obj = %5d.  Value = %08x.\n", id, Gia_ManObj(g, id).Value)
		}
	}
}
func Gia_IsoCollectData(g *Gia_Man_t, objects *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(objects))
	for _, id := range vec.Vec_IntArray(objects) {
		vec.Vec_IntPush(result, int32(Gia_ManObj(g, id).Value))
	}
	return result
}
func Gia_IsoCompareVecs(a *Gia_Man_t, levelsA *Vec_Wec_t, b *Gia_Man_t, levelsB *Vec_Wec_t) {
	fmt.Printf("Graph 1: %d inputs, %d ANDs, %d outputs.\n", Gia_ManCiNum(a), Gia_ManAndNum(a), Gia_ManCoNum(a))
	fmt.Printf("Graph 2: %d inputs, %d ANDs, %d outputs.\n", Gia_ManCiNum(b), Gia_ManAndNum(b), Gia_ManCoNum(b))
	count := min(vec.Vec_WecSize(levelsA), vec.Vec_WecSize(levelsB))
	fmt.Printf("Printing %d shared levels:\n", count)
	for i := 0; i < count; i++ {
		x, y := Gia_IsoCollectData(a, vec.Vec_WecEntry(levelsA, i)), Gia_IsoCollectData(b, vec.Vec_WecEntry(levelsB, i))
		vec.Vec_IntSort(x, 0)
		vec.Vec_IntSort(y, 0)
		common := vec.Vec_IntTwoCountCommon(x, y)
		fmt.Printf("Level = %3d. One = %6d. Two = %6d.  Common = %6d.\n", i, vec.Vec_IntSize(x)-common, vec.Vec_IntSize(y)-common, common)
	}
}
func Gia_Iso4TestTwo(a, b *Gia_Man_t) {
	levelsA, levelsB := Gia_Iso4Gia(a), Gia_Iso4Gia(b)
	Gia_IsoCompareVecs(a, levelsA, b, levelsB)
}
