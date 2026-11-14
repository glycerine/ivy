package gia

import (
	"fmt"
	"slices"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManSimCollect_rec gathers the AND supergate, stopping at complemented edges.
func Gia_ManSimCollect_rec(p *Gia_Man_t, lit int32, result *Vec_Int_t) {
	vec.Vec_IntPush(result, lit)
	obj := Gia_ManObj(p, lit>>1)
	if lit&1 != 0 || Gia_ObjIsCi(obj) != 0 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: simulation supergate expects AND")
	}
	Gia_ManSimCollect_rec(p, 2*Gia_ObjFaninId0(obj, obj.id)+Gia_ObjFaninC0(obj), result)
	Gia_ManSimCollect_rec(p, 2*Gia_ObjFaninId1(obj, obj.id)+Gia_ObjFaninC1(obj), result)
}

func Gia_ManSimCollect(p *Gia_Man_t, obj *Gia_Obj_t, result *Vec_Int_t) {
	vec.Vec_IntClear(result)
	Gia_ManSimCollect_rec(p, 2*obj.id, result)
	entries := vec.Vec_IntArray(result)
	slices.Sort(entries)
	unique := slices.Compact(entries)
	vec.Vec_IntShrink(result, len(unique))
}

// Gia_ManSimDeriveResets counts repeated implications on register inputs.
func Gia_ManSimDeriveResets(p *Gia_Man_t) *Vec_Int_t {
	started := time.Now()
	counts := make([]int32, 2*p.nObjs)
	super := vec.Vec_IntAlloc(1000)
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		ri := Gia_ManRi(p, i)
		fan := Gia_ObjFaninId0(ri, ri.id)
		if fan == 0 {
			continue
		}
		counts[2*fan+Gia_ObjFaninC0(ri)]++
		Gia_ManSimCollect(p, Gia_ManObj(p, fan), super)
		for _, lit := range vec.Vec_IntArray(super) {
			counts[lit]++
		}
	}
	result := vec.Vec_IntStartFull(int(p.nObjs))
	var zeros, ones, piZeros, piOnes int32
	for lit, count := range counts {
		if count < 5 {
			continue
		}
		id := int32(lit >> 1)
		value := int32(1)
		if lit&1 != 0 {
			value = 0
		}
		vec.Vec_IntWriteEntry(result, int(id), value)
		isPi := Gia_ObjIsPi(p, Gia_ManObj(p, id))
		if value == 0 {
			zeros++
			piZeros += isPi
		} else {
			ones++
			piOnes += isPi
		}
	}
	fmt.Printf("Logic0 = %d (%d). Logic1 = %d (%d). Time: %s\n", zeros, piZeros, ones, piOnes, time.Since(started))
	return result
}

func Gia_ManSimSetDefaultParams(p *Gia_ParSim_t) {
	*p = Gia_ParSim_t{NWords: 8, NIters: 32, TimeLimit: 60, iOutFail: -1}
}
