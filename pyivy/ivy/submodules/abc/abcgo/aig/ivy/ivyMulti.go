package ivy

import (
	"fmt"
	"math/bits"
)

const IVY_EVAL_LIMIT = 128

// Ivy_Eva_t records the argument and leaf support represented by a candidate
// node during algebraic rewriting.
type Ivy_Eva_t struct {
	pArg   *Ivy_Obj_t
	Mask   uint32
	Weight int32
}

func ivyMultiVecObj(v *Vec_Ptr_t, i int32) *Ivy_Obj_t {
	obj, ok := Vec_PtrEntry(v, i).(*Ivy_Obj_t)
	if !ok || obj == nil {
		panic("ivy: nil or non-object entry in multi-input vector")
	}
	return obj
}

// Ivy_MultiPlus searches the existing cut and hash table for a cover with
// shared subgraphs. The evaluation workspace is local to the call, unlike the
// original C static array, so concurrent managers do not share mutable state.
func Ivy_MultiPlus(p *Ivy_Man_t, vLeaves, vCone *Vec_Ptr_t, typ Ivy_Type_t, nLimit int32, vSols *Vec_Ptr_t) int32 {
	nLeaves, nCone := Vec_PtrSize(vLeaves), Vec_PtrSize(vCone)
	assert(nLeaves > 2, "nLeaves > 2")
	if nLeaves > 32 || nLeaves+nCone > IVY_EVAL_LIMIT || nLimit <= 0 {
		return 0
	}
	if p == nil || vSols == nil {
		panic("ivy: nil manager or solution vector")
	}

	var evals [IVY_EVAL_LIMIT]Ivy_Eva_t
	var all uint32
	if nLeaves == 32 {
		all = ^uint32(0)
	} else {
		all = uint32(1)<<uint(nLeaves) - 1
	}
	nEvals := int32(0)
	for i := int32(0); i < nLeaves; i++ {
		obj := ivyMultiVecObj(vLeaves, i)
		ivyCheckOperands(p, obj)
		evals[nEvals] = Ivy_Eva_t{pArg: obj, Mask: uint32(1) << uint(nEvals), Weight: 1}
		Ivy_Regular(obj).TravId = nEvals
		nEvals++
	}

	// First annotate every cone node with the union of its leaf supports.
	// Keep the masks separately: the second pass rewrites TravId values for
	// reusable nodes, so indexing back through a fanin's TravId would otherwise
	// read a different entry for later nodes in the cone.
	coneMasks := make([]uint32, int(nCone))
	for i := int32(0); i < nCone; i++ {
		obj := ivyMultiVecObj(vCone, i)
		ivyCheckOperands(p, obj)
		obj.TravId = int32(nEvals + i)
		fanin0 := Ivy_ObjFanin0(obj)
		index0 := fanin0.TravId
		if index0 < 0 || index0 >= nEvals+i {
			panic("ivy: cone is not in topological order or omits a leaf")
		}
		mask := evals[index0].Mask
		if Ivy_ObjIsBuf(obj) == 0 {
			fanin1 := Ivy_ObjFanin1(obj)
			index1 := fanin1.TravId
			if index1 < 0 || index1 >= nEvals+i {
				panic("ivy: cone is not in topological order or omits a leaf")
			}
			mask |= evals[index1].Mask
		}
		coneMasks[i] = mask
		evals[nEvals+i].Mask = mask
	}

	// Save externally referenced cone nodes as reusable candidates. The last
	// node is the root being rewritten and must not be reused as its own result.
	for i := int32(0); i+1 < nCone; i++ {
		obj := ivyMultiVecObj(vCone, i)
		if Ivy_ObjIsBuf(obj) != 0 || Ivy_ObjRefs(obj) == 0 {
			continue
		}
		assert(Ivy_IsComplement(obj) == 0, "Ivy_MultiPlus cone nodes are regular")
		mask := coneMasks[i]
		evals[nEvals] = Ivy_Eva_t{pArg: obj, Mask: mask, Weight: int32(bits.OnesCount32(mask))}
		obj.TravId = nEvals
		nEvals++
	}

	// Enumerate already-existing pair nodes. A candidate must be live, outside
	// the cone, and distinct from all leaves; the latter avoids replacing a
	// trivial input with a supposed shared implementation.
	for i := int32(1); i < nEvals; i++ {
		for k := int32(0); k < i; k++ {
			fan0, fan1 := evals[i], evals[k]
			ghost, ok := ivyMultiGhost(p, fan0.pArg, fan1.pArg, typ)
			if !ok {
				continue
			}
			temp := Ivy_TableLookup(p, ghost)
			if temp == nil || temp.fMarkB != 0 {
				continue
			}
			inLeaves := false
			for x := int32(0); x < nLeaves; x++ {
				if temp == Ivy_Regular(ivyMultiVecObj(vLeaves, x)) {
					inLeaves = true
					break
				}
			}
			if inLeaves {
				continue
			}
			mask := fan0.Mask | fan1.Mask
			weight := fan0.Weight + fan1.Weight
			if fan0.Mask&fan1.Mask != 0 {
				weight = int32(bits.OnesCount32(mask))
			}
			if nEvals >= int32(len(evals)) {
				break
			}
			evals[nEvals] = Ivy_Eva_t{pArg: temp, Mask: mask, Weight: weight}
			// The original C accidentally updated pObj (the last cone node)
			// here. It must mark the discovered candidate so subsequent pairs
			// can use its evaluation record.
			temp.TravId = nEvals
			nEvals++
			if nEvals == int32(len(evals)) || mask == all {
				goto candidatesDone
			}
		}
	}

candidatesDone:
	if Ivy_MultiCover(p, evals[:], nLeaves, nEvals, nLimit, vSols) == 0 {
		return 0
	}
	assert(Vec_PtrSize(vSols) > 0, "Vec_PtrSize(vSols) > 0")
	return 1
}

// ivyMultiGhost returns a hashable prospective operation without asking the
// stricter public ghost constructor to accept degenerate/constant pairs.
func ivyMultiGhost(p *Ivy_Man_t, a, b *Ivy_Obj_t, typ Ivy_Type_t) (*Ivy_Obj_t, bool) {
	if a == nil || b == nil || Ivy_Regular(a) == Ivy_Regular(b) {
		return nil, false
	}
	if typ == IVY_AND && (Ivy_ObjIsConst1(a) != 0 || Ivy_ObjIsConst1(b) != 0) {
		return nil, false
	}
	if Ivy_ObjIsConst1(b) != 0 {
		return nil, false
	}
	return Ivy_ObjCreateGhost(p, a, b, typ, IVY_INIT_NONE), true
}

func Ivy_MultiPrint(_ *Ivy_Man_t, evals []Ivy_Eva_t, nLeaves, nEvals int32) {
	for i := nLeaves; i < nEvals; i++ {
		e := evals[i]
		fmt.Printf("%2d  (id = %5d)  : |", i-nLeaves, Ivy_ObjId(e.pArg))
		for k := int32(0); k < nLeaves; k++ {
			if e.Mask&(uint32(1)<<uint(k)) != 0 {
				fmt.Print("+")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Printf("|  Lev = %d.\n", Ivy_ObjLevel(e.pArg))
	}
}

func Ivy_MultiWeight(mask uint32, maskWeight int32, found uint32) int32 {
	assert(mask&^found != 0, "mask &^ found != 0")
	if mask&found == 0 {
		return maskWeight
	}
	return int32(bits.OnesCount32(mask &^ found))
}

func Ivy_MultiCover(_ *Ivy_Man_t, evals []Ivy_Eva_t, nLeaves, nEvals, nLimit int32, solutions *Vec_Ptr_t) int32 {
	var all uint32
	if nLeaves == 32 {
		all = ^uint32(0)
	} else {
		all = uint32(1)<<uint(nLeaves) - 1
	}
	Vec_PtrClear(solutions)
	var found uint32
	for i := int32(0); i < nLimit; i++ {
		bestIndex := int32(-1)
		var bestWeight, bestLevel int32
		for k := nEvals - 1; k >= 0; k-- {
			e := &evals[k]
			if e.Mask&^found == 0 {
				continue
			}
			weight := Ivy_MultiWeight(e.Mask, e.Weight, found)
			level := Ivy_ObjLevel(Ivy_Regular(e.pArg))
			if bestIndex < 0 || weight > bestWeight || weight == bestWeight && level < bestLevel {
				bestIndex, bestWeight, bestLevel = k, weight, level
			}
		}
		assert(bestIndex >= 0, "bestIndex >= 0")
		best := &evals[bestIndex]
		if bestWeight == 1 && bestIndex >= nLeaves {
			remaining := best.Mask &^ found
			leaf := int32(bits.TrailingZeros32(remaining))
			assert(leaf < nLeaves, "leaf < nLeaves")
			bestIndex = leaf
			best = &evals[bestIndex]
		}
		Vec_PtrPush(solutions, best.pArg)
		found |= best.Mask
		if found == all {
			return 1
		}
	}
	return 0
}
