package acec

// Full-adder cut enumeration from src/proof/acec/acecFadds.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Dtc_ManCutMerge(p *Gia_Man_t, id int32, list0, list1 []int32, cuts, xor2, xor3, maj *Vec_Int_t) {
	vec.Vec_IntFill(cuts, 2, 1)
	vec.Vec_IntPush(cuts, id)
	off0 := 1
	for i := int32(0); i < list0[0]; i++ {
		n0 := list0[off0]
		a := list0[off0 : off0+int(n0)+1]
		off0 += int(n0) + 1
		off1 := 1
		for j := int32(0); j < list1[0]; j++ {
			n1 := list1[off1]
			b := list1[off1 : off1+int(n1)+1]
			off1 += int(n1) + 1
			var cut [4]int32
			if Dtc_ManCutMergeOne(a, b, cut[:]) == 0 || Dtc_ManCutCheckEqual(cuts, cut[:]) != 0 {
				continue
			}
			vec.Vec_IntAddToEntry(cuts, 0, 1)
			for _, v := range cut[:1+cut[0]] {
				vec.Vec_IntPush(cuts, v)
			}
			typ := Dtc_ObjComputeTruth(p, id, cut[:], nil)
			if cut[0] == 2 {
				if typ != 0 && typ != 3 {
					panic("acec: unexpected two-input cut type")
				}
				if typ == 3 {
					vec.Vec_IntPushThree(xor2, cut[1], cut[2], id)
				}
				continue
			}
			if cut[0] != 3 || typ == 0 {
				continue
			}
			dst := maj
			if typ == 1 {
				dst = xor3
			}
			for _, v := range cut[1:4] {
				vec.Vec_IntPush(dst, v)
			}
			vec.Vec_IntPush(dst, id)
		}
	}
}
func Dtc_ManComputeCuts(p *Gia_Man_t, pxor2, pxor3, pmaj **Vec_Int_t, verbose int32) {
	if !acecCanDetectArithmetic(p) {
		panic("acec: cut detection requires AND/XOR nodes without native MUXes or buffers")
	}
	temp := vec.Vec_IntAlloc(1000)
	xor2, xor3, maj := vec.Vec_IntAlloc(int(gia.Gia_ManAndNum(p))), vec.Vec_IntAlloc(int(gia.Gia_ManAndNum(p))), vec.Vec_IntAlloc(int(gia.Gia_ManAndNum(p)))
	cuts := vec.Vec_IntAlloc(30 * int(gia.Gia_ManAndNum(p)))
	vec.Vec_IntFill(cuts, int(gia.Gia_ManObjNum(p)), 0)
	gia.Gia_ManCleanValue(p)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
		vec.Vec_IntWriteEntry(cuts, int(id), int32(vec.Vec_IntSize(cuts)))
		vec.Vec_IntPushThree(cuts, 1, 1, id)
	}
	total := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		list := vec.Vec_IntArray(cuts)
		a, b := int(vec.Vec_IntEntry(cuts, int(gia.Gia_ObjFaninId0(o, id)))), int(vec.Vec_IntEntry(cuts, int(gia.Gia_ObjFaninId1(o, id))))
		Dtc_ManCutMerge(p, id, list[a:], list[b:], temp, xor2, xor3, maj)
		vec.Vec_IntWriteEntry(cuts, int(id), int32(vec.Vec_IntSize(cuts)))
		vec.Vec_IntAppend(cuts, temp)
		total += int(vec.Vec_IntEntry(temp, 0))
	}
	if verbose != 0 {
		n := float64(gia.Gia_ManAndNum(p))
		fmt.Printf("Nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.\n", gia.Gia_ManAndNum(p), total, float64(total)/n, float64(vec.Vec_IntSize(cuts))/n)
	}
	if pxor2 != nil {
		*pxor2 = xor2
	}
	*pxor3 = xor3
	*pmaj = maj
}
func Dtc_ManFindCommonCuts(p *Gia_Man_t, xors, majs *Vec_Int_t) *Vec_Int_t {
	a, b := vec.Vec_IntArray(xors), vec.Vec_IntArray(majs)
	result := vec.Vec_IntAlloc(1000)
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		order := Dtc_ManCompare(a[i:i+4], b[j:j+4])
		if order == 0 {
			for _, x := range a[i : i+4] {
				vec.Vec_IntPush(result, x)
			}
			vec.Vec_IntPush(result, b[j+3])
			i += 4
			j += 4
		} else if order < 0 {
			i += 4
		} else {
			j += 4
		}
	}
	return result
}
func Dtc_ManPrintFadds(adds *Vec_Int_t) {
	for i := 0; i < vec.Vec_IntSize(adds)/5; i++ {
		fmt.Printf("%6d : %6d %6d %6d  ->  %6d %6d\n", i, vec.Vec_IntEntry(adds, 5*i), vec.Vec_IntEntry(adds, 5*i+1), vec.Vec_IntEntry(adds, 5*i+2), vec.Vec_IntEntry(adds, 5*i+3), vec.Vec_IntEntry(adds, 5*i+4))
		if i == 100 {
			fmt.Println("Skipping other FADDs.")
			break
		}
	}
}
func Dtc_ManCompare(a, b []int32) int32 {
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
func Dtc_ManCompare2(a, b []int32) int32 {
	if a[4] < b[4] {
		return -1
	}
	if a[4] > b[4] {
		return 1
	}
	return 0
}
func Gia_ManDetectFullAdders(p *Gia_Man_t, verbose int32, pxor2 **Vec_Int_t) *Vec_Int_t {
	var xors, majs *Vec_Int_t
	Dtc_ManComputeCuts(p, pxor2, &xors, &majs, verbose)
	for _, v := range []*Vec_Int_t{xors, majs} {
		raw := vec.Vec_IntArray(v)
		records := make([][4]int32, len(raw)/4)
		for i := range records {
			copy(records[i][:], raw[4*i:4*i+4])
		}
		sort.SliceStable(records, func(i, j int) bool { return Dtc_ManCompare(records[i][:], records[j][:]) < 0 })
		for i := range records {
			copy(raw[4*i:4*i+4], records[i][:])
		}
	}
	adds := Dtc_ManFindCommonCuts(p, xors, majs)
	raw := vec.Vec_IntArray(adds)
	records := make([][5]int32, len(raw)/5)
	for i := range records {
		copy(records[i][:], raw[5*i:5*i+5])
	}
	sort.SliceStable(records, func(i, j int) bool { return Dtc_ManCompare2(records[i][:], records[j][:]) < 0 })
	for i := range records {
		copy(raw[5*i:5*i+5], records[i][:])
	}
	if verbose != 0 {
		fmt.Printf("XOR3 cuts = %d.  MAJ cuts = %d.  Full-adders = %d.\n", vec.Vec_IntSize(xors)/4, vec.Vec_IntSize(majs)/4, vec.Vec_IntSize(adds)/5)
		Dtc_ManPrintFadds(adds)
	}
	return adds
}
