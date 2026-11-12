package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func giaSimConstants(words []uint64) (bool, bool) {
	zero, one := true, true
	for _, w := range words {
		zero = zero && w == 0
		one = one && w == ^uint64(0)
	}
	return zero, one
}
func Gia_ManSimPatHashPatterns(p *Gia_Man_t, n int32, sims *Vec_Wrd_t, c0, c1 *int32) int32 {
	if n <= 0 || vec.Vec_WrdSize(sims) != int(n*p.nObjs) {
		panic("gia: invalid simulation dimensions")
	}
	store := vec.Vec_MemAlloc(int(n), 12)
	vec.Vec_MemHashAlloc(store, 1<<12)
	defer vec.Vec_MemFree(store)
	data := vec.Vec_WrdArray(sims)
	for _, obj := range p.pObjs {
		if Gia_ObjIsCand(obj) == 0 {
			continue
		}
		words := data[obj.id*n : (obj.id+1)*n]
		zero, one := giaSimConstants(words)
		if c0 != nil && zero {
			*c0++
		}
		if c1 != nil && one {
			*c1++
		}
		vec.Vec_MemHashInsert(store, words)
	}
	return int32(vec.Vec_MemEntryNum(store))
}
func Gia_ManSimPatGenMiter(p *Gia_Man_t, sims *Vec_Wrd_t) *Gia_Man_t {
	if p.nObjs == 0 || vec.Vec_WrdSize(sims)%int(p.nObjs) != 0 || vec.Vec_WrdSize(sims) == 0 {
		panic("gia: invalid simulation dimensions")
	}
	n := int32(vec.Vec_WrdSize(sims) / int(p.nObjs))
	data := vec.Vec_WrdArray(sims)
	dst := Gia_ManStart(p.nObjs + Gia_ManCoNum(p))
	Gia_ManHashStart(dst)
	Gia_ManConst0(p).Value = 0
	for _, id := range vec.Vec_IntArray(p.vCis) {
		Gia_ManObj(p, id).Value = uint32(Gia_ManAppendCi(dst))
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		zero, one := giaSimConstants(data[obj.id*n : (obj.id+1)*n])
		if zero {
			Gia_ManAppendCo(dst, int32(obj.Value))
		}
		if one {
			Gia_ManAppendCo(dst, int32(obj.Value)^1)
		}
	}
	Gia_ManHashStop(dst)
	return dst
}
func Gia_ManSimProfile(p *Gia_Man_t) {
	sims := Gia_ManSimPatSim(p)
	n := int32(vec.Vec_WrdSize(sims)) / p.nObjs
	var c0, c1 int32
	unique := Gia_ManSimPatHashPatterns(p, n, sims, &c0, &c1)
	count := Gia_ManCandNum(p)
	percentage := float64(0)
	if count != 0 {
		percentage = 100 * float64(unique) / float64(count)
	}
	fmt.Printf("Simulating %d patterns leads to %d unique objects (%.2f %% out of %d), Const0 = %d. Const1 = %d.\n", 64*n, unique, percentage, count, c0, c1)
}
