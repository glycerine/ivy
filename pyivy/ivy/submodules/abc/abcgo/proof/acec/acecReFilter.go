package acec

// Adder cut filtering and diagnostics from src/proof/acec/acecRe.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Ree_CollectInsiders_rec(p *Gia_Man_t, id int32, visited, insiders *Vec_Bit_t) {
	if vec.Vec_BitEntry(visited, int(id)) != 0 {
		return
	}
	vec.Vec_BitWriteEntry(visited, int(id), 1)
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: unmarked adder input")
	}
	Ree_CollectInsiders_rec(p, gia.Gia_ObjFaninId0p(p, o), visited, insiders)
	Ree_CollectInsiders_rec(p, gia.Gia_ObjFaninId1p(p, o), visited, insiders)
	vec.Vec_BitWriteEntry(insiders, int(id), 1)
}
func Ree_CollectInsiders(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Bit_t {
	visited, inside := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		if vec.Vec_IntEntry(adds, 6*i+2) == 0 {
			continue
		}
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(visited, int(vec.Vec_IntEntry(adds, 6*i+k)), 1)
		}
		for k := 3; k < 5; k++ {
			Ree_CollectInsiders_rec(p, vec.Vec_IntEntry(adds, 6*i+k), visited, inside)
		}
	}
	return inside
}
func Ree_ManRemoveTrivial(p *Gia_Man_t, adds *Vec_Int_t) {
	p.Refs = nil
	gia.Gia_ManCreateRefs(p)
	data := vec.Vec_IntArray(adds)
	k := 0
	for i := 0; i < len(data); i += 6 {
		if data[i+2] == 0 {
			x, m := gia.Gia_ManObj(p, data[i+3]), gia.Gia_ManObj(p, data[i+4])
			if (m == gia.Gia_ObjFanin0(x) || m == gia.Gia_ObjFanin1(x)) && gia.Gia_ObjRefNum(p, m) == 1 {
				continue
			}
		}
		copy(data[6*k:6*k+6], data[i:i+6])
		k++
	}
	vec.Vec_IntShrink(adds, 6*k)
}
func Ree_ManRemoveContained(p *Gia_Man_t, adds *Vec_Int_t) {
	inside := Ree_CollectInsiders(p, adds)
	data := vec.Vec_IntArray(adds)
	k := 0
	for i := 0; i < len(data); i += 6 {
		if data[i+2] == 0 && vec.Vec_BitEntry(inside, int(data[i+3])) != 0 && vec.Vec_BitEntry(inside, int(data[i+4])) != 0 {
			continue
		}
		copy(data[6*k:6*k+6], data[i:i+6])
		k++
	}
	vec.Vec_IntShrink(adds, 6*k)
}
func Ree_ManCountFadds(adds *Vec_Int_t) int32 {
	var n int32
	for i := 2; i < vec.Vec_IntSize(adds); i += 6 {
		if vec.Vec_IntEntry(adds, i) != 0 {
			n++
		}
	}
	return n
}
func Ree_ManPrintAdders(adds *Vec_Int_t, verbose int32) {
	if verbose == 0 {
		return
	}
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		fmt.Printf("%6d : %6d %6d %6d    ->  %6d %6d  (%d)\n", i, vec.Vec_IntEntry(adds, 6*i), vec.Vec_IntEntry(adds, 6*i+1), vec.Vec_IntEntry(adds, 6*i+2), vec.Vec_IntEntry(adds, 6*i+3), vec.Vec_IntEntry(adds, 6*i+4), vec.Vec_IntEntry(adds, 6*i+5))
	}
}
func Ree_ManComputeCutsTest(p *Gia_Man_t) {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	n := Ree_ManCountFadds(adds)
	Ree_ManPrintAdders(adds, 1)
	fmt.Printf("Detected %d FAs and %d HAs. Time = %v\n", n, int32(vec.Vec_IntSize(adds)/6)-n, time.Since(start))
}
