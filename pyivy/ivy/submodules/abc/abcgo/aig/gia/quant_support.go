package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManQuantInfoId(p *Gia_Man_t, id int32) []word {
	start := id * p.nSuppWords
	return vec.Vec_WrdArray(p.vSuppWords)[start : start+p.nSuppWords]
}
func Gia_ManQuantInfo(p *Gia_Man_t, obj *Gia_Obj_t) []word { return Gia_ManQuantInfoId(p, obj.id) }
func Gia_ObjCopyGetTwoArray(p *Gia_Man_t, id int32, out []int32) {
	out[0] = vec.Vec_IntEntry(&p.vCopiesTwo, int(2*id))
	out[1] = vec.Vec_IntEntry(&p.vCopiesTwo, int(2*id+1))
}
func Gia_ObjCopySetTwoArray(p *Gia_Man_t, id int32, in []int32) {
	vec.Vec_IntWriteEntry(&p.vCopiesTwo, int(2*id), in[0])
	vec.Vec_IntWriteEntry(&p.vCopiesTwo, int(2*id+1), in[1])
}
func Gia_ManQuantVerify_rec(p *Gia_Man_t, id, ci int32) int32 {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return giaBool(Gia_ObjCioId(obj) == ci)
	}
	if id == 0 {
		return 0
	}
	return giaBool(Gia_ManQuantVerify_rec(p, Gia_ObjFaninId0(obj, id), ci) != 0 || Gia_ManQuantVerify_rec(p, Gia_ObjFaninId1(obj, id), ci) != 0)
}
func Gia_ManQuantVerify(p *Gia_Man_t, id int32) {
	for i, ci := range vec.Vec_IntArray(&p.vSuppVars) {
		Gia_ManIncrementTravId(p)
		if Gia_ManQuantCheckSupp(p, id, int32(i)) != Gia_ManQuantVerify_rec(p, id, ci) {
			fmt.Printf("Mismatch at node %d related to CI %d (%d).\n", id, ci, Gia_ManQuantCheckSupp(p, id, int32(i)))
		}
	}
}
func Gia_ManQuantSetSuppStart(p *Gia_Man_t) {
	if p.nObjs != 1 || p.vSuppWords != nil || vec.Vec_IntSize(&p.vSuppVars) != 0 {
		panic("gia: invalid support initialization")
	}
	p.iSuppPi = 0
	p.nSuppWords = 1
	p.vSuppWords = vec.Vec_WrdAlloc(1000)
	vec.Vec_WrdPush(p.vSuppWords, 0)
}
func Gia_ManQuantSetSuppZero(p *Gia_Man_t) {
	for i := int32(0); i < p.nSuppWords; i++ {
		vec.Vec_WrdPush(p.vSuppWords, 0)
	}
	if vec.Vec_WrdSize(p.vSuppWords) != int(p.nSuppWords*p.nObjs) {
		panic("gia: invalid support size")
	}
}
func Gia_ManQuantSetSuppCi(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsCi(obj) == 0 || p.vSuppWords == nil {
		panic("gia: invalid support input")
	}
	if p.iSuppPi == 64*p.nSuppWords {
		old := vec.Vec_WrdArray(p.vSuppWords)
		replacement := vec.Vec_WrdAlloc(len(old) + int(p.nObjs))
		for i := 0; i < len(old); i += int(p.nSuppWords) {
			for _, x := range old[i : i+int(p.nSuppWords)] {
				vec.Vec_WrdPush(replacement, x)
			}
			vec.Vec_WrdPush(replacement, 0)
		}
		p.vSuppWords = replacement
		p.nSuppWords++
	}
	if p.iSuppPi != int32(vec.Vec_IntSize(&p.vSuppVars)) {
		panic("gia: inconsistent support variables")
	}
	vec.Vec_IntPush(&p.vSuppVars, Gia_ObjCioId(obj))
	info := Gia_ManQuantInfo(p, obj)
	info[p.iSuppPi>>6] |= uint64(1) << uint(p.iSuppPi&63)
	p.iSuppPi++
}
func Gia_ManQuantCheckSupp(p *Gia_Man_t, id, bit int32) int32 {
	info := Gia_ManQuantInfoId(p, id)
	return giaBool(info[bit>>6]&(uint64(1)<<uint(bit&63)) != 0)
}
func Gia_ManQuantUpdateCiSupp(p *Gia_Man_t, id int32) {
	info := Gia_ManQuantInfoId(p, id)
	var any bool
	for _, x := range info {
		any = any || x != 0
	}
	if !any {
		Gia_ManQuantSetSuppCi(p, Gia_ManObj(p, id))
	}
}
func Gia_ManQuantCheckOverlap(p *Gia_Man_t, id int32) int32 {
	a, b := Gia_ManQuantInfoId(p, id), Gia_ManQuantInfoId(p, 0)
	for i := range a {
		if a[i]&b[i] != 0 {
			return 1
		}
	}
	return 0
}
func Gia_ManQuantMarkUsedCis(p *Gia_Man_t, keep func(any, int32) int32, data any) {
	info := Gia_ManQuantInfoId(p, 0)
	clear(info)
	for i, ci := range vec.Vec_IntArray(&p.vSuppVars) {
		if keep(data, ci) == 0 {
			info[i>>6] |= uint64(1) << uint(i&63)
		}
	}
}
func Gia_ManQuantCountUsed_rec(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return 0
	}
	count := int32(1)
	for _, fanin := range []int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)} {
		if Gia_ManQuantCheckSupp(p, fanin, p.iSuppPi) != 0 {
			count += Gia_ManQuantCountUsed_rec(p, fanin)
		}
	}
	return count
}
func Gia_ManQuantCountUsed(p *Gia_Man_t, id int32) int32 {
	Gia_ManIncrementTravId(p)
	return Gia_ManQuantCountUsed_rec(p, id)
}
