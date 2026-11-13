package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Gia_GlaCollectAssigned(p *Gia_Man_t, classes *Vec_Int_t) *Vec_Int_t {
	seen := make(map[int32]bool)
	add := func(id int32) { seen[id] = true }
	for i, value := range vec.Vec_IntArray(classes) {
		if value == 0 {
			continue
		}
		if value < 0 {
			panic("gia: invalid gate class")
		}
		obj := Gia_ManObj(p, int32(i))
		add(obj.id)
		if Gia_ObjIsAnd(obj) != 0 {
			add(Gia_ObjFanin0(obj).id)
			add(Gia_ObjFanin1(obj).id)
			if control, _, _, kind := p.Gate(obj.id); kind == 4 {
				add(control >> 1)
			}
		} else if Gia_ObjIsRo(p, obj) != 0 {
			add(Gia_ObjFanin0(Gia_ObjRoToRi(p, obj)).id)
		} else if Gia_ObjIsConst0(obj) == 0 {
			panic("gia: invalid classified object")
		}
	}
	ids := make([]int, 0, len(seen))
	for id := range seen {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	result := vec.Vec_IntAlloc(len(ids))
	for _, id := range ids {
		vec.Vec_IntPush(result, int32(id))
	}
	return result
}
func Gia_ManGlaCollect(p *Gia_Man_t, classes *Vec_Int_t, pis, ppis, flops, nodes **Vec_Int_t) {
	if Gia_ManPoNum(p) != 1 || vec.Vec_IntSize(classes) != int(p.nObjs) {
		panic("gia: invalid gate abstraction")
	}
	for _, out := range []**Vec_Int_t{pis, ppis, flops, nodes} {
		if out != nil {
			*out = vec.Vec_IntAlloc(100)
		}
	}
	for _, id := range vec.Vec_IntArray(Gia_GlaCollectAssigned(p, classes)) {
		obj := Gia_ManObj(p, id)
		var target **Vec_Int_t
		if Gia_ObjIsPi(p, obj) != 0 {
			target = pis
		} else if vec.Vec_IntEntry(classes, int(id)) == 0 {
			target = ppis
		} else if Gia_ObjIsRo(p, obj) != 0 {
			target = flops
		} else if Gia_ObjIsAnd(obj) != 0 {
			target = nodes
		} else if Gia_ObjIsConst0(obj) == 0 {
			panic("gia: invalid abstraction member")
		}
		if target != nil {
			vec.Vec_IntPush(*target, id)
		}
	}
}
func Gia_ManPrintFlopClasses(p *Gia_Man_t) {
	if p.vFlopClasses == nil {
		return
	}
	if vec.Vec_IntSize(p.vFlopClasses) != int(Gia_ManRegNum(p)) {
		fmt.Println("Gia_ManPrintFlopClasses(): The number of flop map entries differs from the number of flops.")
		return
	}
	zero, one := vec.Vec_IntCountEntry(p.vFlopClasses, 0), vec.Vec_IntCountEntry(p.vFlopClasses, 1)
	fmt.Printf("Flop-level abstraction:  Excluded FFs = %d  Included FFs = %d  (%.2f %%) ", zero, one, 100*float64(one)/float64(zero+one+1))
	if zero+one < int(Gia_ManRegNum(p)) {
		fmt.Print("and there are other FF classes...")
	}
	fmt.Println()
}
func Gia_ManPrintGateClasses(p *Gia_Man_t) {
	if p.vGateClasses == nil {
		return
	}
	if vec.Vec_IntSize(p.vGateClasses) != int(p.nObjs) {
		fmt.Println("Gia_ManPrintGateClasses(): The number of flop map entries differs from the number of flops.")
		return
	}
	var pis, ppis, flops, nodes *Vec_Int_t
	Gia_ManGlaCollect(p, p.vGateClasses, &pis, &ppis, &flops, &nodes)
	a, b, c, d := vec.Vec_IntSize(pis), vec.Vec_IntSize(ppis), vec.Vec_IntSize(flops), vec.Vec_IntSize(nodes)
	total := 1 + c + d
	fmt.Printf("Gate-level abstraction:  PI = %d  PPI = %d  FF = %d (%.2f %%)  AND = %d (%.2f %%)  Obj = %d (%.2f %%)\n", a, b, c, 100*float64(c)/float64(Gia_ManRegNum(p)+1), d, 100*float64(d)/float64(Gia_ManAndNum(p)+1), total, 100*float64(total)/float64(Gia_ManRegNum(p)+Gia_ManAndNum(p)+1))
}
func Gia_ManPrintObjClasses(p *Gia_Man_t) {
	if p.vObjClasses == nil {
		return
	}
	data := vec.Vec_IntArray(p.vObjClasses)
	frames := int(data[0])
	if frames < 0 || frames+1 >= len(data) || int(data[frames+1]) != len(data) {
		panic("gia: invalid object classes")
	}
	nBits := util.Abc_Base2Log(uint32(p.nObjs))
	mask := int32((1 << uint(nBits)) - 1)
	if p.nObjs > mask {
		panic("gia: invalid object mask")
	}
	fmt.Println("Frame   Core   F0   F1   F2   F3 ...")
	for i := 0; i < frames; i++ {
		start, end := int(data[i+1]), int(data[i+2])
		counts := make([]int, frames+1)
		if start < 0 || start > end || end > len(data) {
			panic("gia: invalid frame offsets")
		}
		for _, entry := range data[start:end] {
			id, frame := entry&mask, entry>>uint(nBits)
			if id >= p.nObjs || frame < 0 || int(frame) >= frames {
				panic("gia: invalid frame object")
			}
			counts[frame+1]++
			counts[0]++
		}
		fmt.Printf("%3d :%7d", i, counts[0])
		if i >= 10 {
			for k := 0; k < 4; k++ {
				fmt.Printf("%5d", counts[k+1])
			}
			fmt.Print("  ...")
			for k := i - 4; k <= i; k++ {
				fmt.Printf("%5d", counts[k+1])
			}
		} else {
			for k := 0; k <= i; k++ {
				fmt.Printf("%5d", counts[k+1])
			}
		}
		fmt.Println()
	}
}
