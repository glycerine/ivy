package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_ClassesPrepareRehash groups candidates by simulation hash, then refines collisions.
func Ssw_ClassesPrepareRehash(p *Ssw_Cla_t, candidates *vec.Vec_Ptr_t, constOnly int32) int32 {
	tableSize := int(util.Abc_PrimeCudd(uint32(vec.Vec_PtrSize(candidates) / 2)))
	table := make([][]aig.Edge, tableSize)
	p.nCands1 = 0
	for i := 0; i < vec.Vec_PtrSize(candidates); i++ {
		obj := vec.Vec_PtrEntry(candidates, i).(aig.Edge)
		if p.pClassSizes[obj.NodeID()] != 0 {
			panic("ssw: candidate already has a class")
		}
		aig.Aig_ObjClearRepr(p.pAig, obj)
		if p.pFuncNodeIsConst(p.pManData, obj) != 0 {
			Ssw_ObjSetConst1Cand(p.pAig, obj)
			p.nCands1++
			continue
		}
		if constOnly != 0 {
			continue
		}
		bucket := int(p.pFuncNodeHash(p.pManData, obj) % uint32(tableSize))
		table[bucket] = append(table[bucket], obj)
	}
	for _, class := range table {
		if len(class) < 2 {
			continue
		}
		repr := class[0]
		for _, obj := range class[1:] {
			aig.Aig_ObjSetRepr(p.pAig, obj, repr)
		}
		Ssw_ObjAddClass(p, repr, class, int32(len(class)))
	}
	return Ssw_ClassesRefine(p, 1)
}
