package cec

import "github.com/glycerine/abcgo/aig/gia"

// Gia_TransferMappedClasses transfers equivalences from a partition to the
// original object's IDs. A map entry is required for every class member.
func Gia_TransferMappedClasses(part *gia.Gia_Man_t, mapBack, reprs []int32) int32 {
	if !part.HasRepresentatives() {
		return 0
	}
	var count int32
	for id := int32(0); id < gia.Gia_ManObjNum(part); id++ {
		repr := gia.Gia_ObjRepr(part, id)
		if repr == gia.GIA_VOID {
			continue
		}
		a, b := mapBack[id], mapBack[repr]
		if a == b {
			continue
		}
		if a < b {
			reprs[b] = a
		} else {
			reprs[a] = b
		}
		count++
	}
	return count
}

func Gia_ManFindRepr_rec(reprs []int32, id int32) int32 {
	if reprs[id] == 0 {
		return 0
	}
	if reprs[id] == -1 {
		return id
	}
	return Gia_ManFindRepr_rec(reprs, reprs[id])
}

// Gia_ManNormalizeEquivalences resolves representative chains and builds
// the linked lists used by the other GIA class routines.
func Gia_ManNormalizeEquivalences(p *gia.Gia_Man_t, reprs []int32) {
	if p.HasRepresentatives() {
		panic("cec: representatives already exist")
	}
	p.StartRepresentatives()
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		if reprs[id] != -1 {
			gia.Gia_ObjSetRepr(p, id, Gia_ManFindRepr_rec(reprs, id))
		}
	}
	for id, next := range gia.Gia_ManDeriveNexts(p) {
		p.SetNext(int32(id), next)
	}
}
