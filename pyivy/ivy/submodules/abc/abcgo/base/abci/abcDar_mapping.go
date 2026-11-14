package abci

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge

// Abc_NtkDarObjLit returns the AIG literal stored by Abc_NtkToDar for obj.
// The bool is false for objects which were not mapped into that AIG.
func Abc_NtkDarObjLit(obj *Abc_Obj_t) (aig.Edge, bool) {
	if obj == nil || obj.Temp < 0 {
		return 0, false
	}
	return aig.Edge(uint32(obj.Temp)), true
}

// Abc_NtkDarMappedObjLit resolves an original network object's CopyEdge
// through the strashed network's Temp mapping left by Abc_NtkToDar. Both the
// copy-edge phase and the phase in the AIG literal are preserved.
func Abc_NtkDarMappedObjLit(obj *Abc_Obj_t) (aig.Edge, bool) {
	if obj == nil {
		return 0, false
	}
	phase := false
	seen := make(map[*Abc_Obj_t]struct{})
	for {
		if _, duplicate := seen[obj]; duplicate {
			return 0, false
		}
		seen[obj] = struct{}{}
		copy := abc.Abc_ObjCopyEdge(obj)
		if copy.Object == nil {
			break
		}
		phase = phase != copy.Complement
		obj = copy.Object
	}
	lit, ok := Abc_NtkDarObjLit(obj)
	if !ok {
		return 0, false
	}
	if phase {
		lit = lit.Not()
	}
	return lit, true
}
