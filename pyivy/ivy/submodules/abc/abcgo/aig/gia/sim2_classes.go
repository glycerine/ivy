package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_Sim2ClassCreate(g *Gia_Man_t, class *Vec_Int_t) {
	ids := vec.Vec_IntArray(class)
	if len(ids) == 0 {
		panic("gia: empty equivalence class")
	}
	head := ids[0]
	Gia_ObjSetRepr(g, head, GIA_VOID)
	for i := 1; i < len(ids); i++ {
		if ids[i] <= head {
			panic("gia: unordered equivalence class")
		}
		Gia_ObjSetRepr(g, ids[i], head)
		Gia_ObjSetNext(g, ids[i-1], ids[i])
	}
	Gia_ObjSetNext(g, ids[len(ids)-1], 0)
}
func Gia_Sim2ClassRefineOne(p *Gia_Sim2_t, id int32) int32 {
	vec.Vec_IntClear(p.vClassOld)
	vec.Vec_IntClear(p.vClassNew)
	vec.Vec_IntPush(p.vClassOld, id)
	head := Gia_ManObj(p.pAig, id)
	sim := Gia_Sim2Data(p, id)
	for node := Gia_ObjNext(p.pAig, id); node != 0; node = Gia_ObjNext(p.pAig, node) {
		obj := Gia_ManObj(p.pAig, node)
		if Gia_Sim2CompareEqual(sim, Gia_Sim2Data(p, node), p.nWords, int32(head.fPhase^obj.fPhase)) != 0 {
			vec.Vec_IntPush(p.vClassOld, node)
		} else {
			vec.Vec_IntPush(p.vClassNew, node)
		}
	}
	if vec.Vec_IntSize(p.vClassNew) == 0 {
		return 0
	}
	Gia_Sim2ClassCreate(p.pAig, p.vClassOld)
	Gia_Sim2ClassCreate(p.pAig, p.vClassNew)
	if vec.Vec_IntSize(p.vClassNew) > 1 {
		return 1 + Gia_Sim2ClassRefineOne(p, vec.Vec_IntEntry(p.vClassNew, 0))
	}
	return 1
}
func Gia_Sim2HashKey(sim []uint32, n, size int32) int32 {
	primes := [16]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	var hash uint32
	for i := int32(0); i < n; i++ {
		word := sim[i]
		if sim[0]&1 != 0 {
			word = ^word
		}
		hash ^= word * primes[i&15]
	}
	return int32(hash % uint32(size))
}
func Gia_Sim2ProcessRefined(p *Gia_Sim2_t, refined *Vec_Int_t) {
	ids := vec.Vec_IntArray(refined)
	if len(ids) == 0 {
		return
	}
	size := util.Abc_PrimeCudd(uint32(1000 + len(ids)/3))
	table := make([]int32, size)
	for _, id := range ids {
		key := Gia_Sim2HashKey(Gia_Sim2Data(p, id), p.nWords, size)
		previous := table[key]
		if previous == 0 {
			if Gia_ObjRepr(p.pAig, id) != 0 || Gia_ObjNext(p.pAig, id) != 0 {
				panic("gia: invalid constant candidate")
			}
			Gia_ObjSetRepr(p.pAig, id, GIA_VOID)
		} else {
			Gia_ObjSetNext(p.pAig, previous, id)
			repr := Gia_ObjRepr(p.pAig, previous)
			if repr == GIA_VOID {
				repr = previous
			}
			Gia_ObjSetRepr(p.pAig, id, repr)
		}
		table[key] = id
	}
}
func Gia_Sim2InfoRefineEquivs(p *Gia_Sim2_t) {
	refined := vec.Vec_IntAlloc(100)
	for id := int32(1); id < p.pAig.nObjs; id++ {
		if Gia_ObjIsConst(p.pAig, id) != 0 && Gia_Sim2CompareZero(Gia_Sim2Data(p, id), p.nWords, int32(Gia_ManObj(p.pAig, id).fPhase)) == 0 {
			vec.Vec_IntPush(refined, id)
		}
	}
	Gia_Sim2ProcessRefined(p, refined)
	for id := int32(1); id < p.pAig.nObjs; id++ {
		if Gia_ObjIsHead(p.pAig, id) != 0 {
			Gia_Sim2ClassRefineOne(p, id)
		}
	}
}
