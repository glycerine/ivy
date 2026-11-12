package ivy

import (
	"fmt"
	"math/bits"
)

// Ivy_Eval_t describes a candidate pair operation by the input support it
// covers, its overlap cost, level estimate, and fanin indices in the
// evaluation graph. Go integers avoid the silent truncation of the C bitfields.
type Ivy_Eval_t struct {
	Mask, Weight, Cost, Level uint32
	Fan0, Fan1                int
}

func Ivy_Multi(p *Ivy_Man_t, initial []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if p == nil || (typ != IVY_AND && typ != IVY_EXOR) || nArgs <= 0 || int64(nArgs) > int64(len(initial)) {
		panic("ivy: invalid multi-input operation")
	}
	args := append([]*Ivy_Obj_t(nil), initial[:nArgs]...)
	for _, arg := range args {
		ivyCheckOperands(p, arg)
	}
	if nArgs == 1 {
		return args[0]
	}
	if nArgs == 2 {
		return Ivy_Oper(p, args[0], args[1], typ)
	}
	// ABC's search records at most 15 candidate arguments. Larger input lists
	// still have a correct and level-aware balanced implementation.
	if nArgs >= 15 || nArgs > 32 {
		Ivy_MultiSort(args, nArgs)
		return Ivy_MultiBalance_rec(p, args, nArgs, typ)
	}

	evals := make([]Ivy_Eval_t, 15+15*14/2)
	for i, arg := range args {
		evals[i] = Ivy_Eval_t{Mask: uint32(1) << uint(i), Weight: 1, Level: uint32(Ivy_ObjLevel(arg))}
	}
	bestIndex := 0
	nArgsNew := len(args)
	for i := 1; i < nArgsNew; i++ {
		for k := 0; k < i; k++ {
			ghost, ok := ivyMultiGhost(p, args[k], args[i], typ)
			if !ok {
				continue
			}
			temp := Ivy_TableLookup(p, ghost)
			if temp == nil {
				continue
			}
			eval := Ivy_Eval_t{
				Mask:   evals[k].Mask | evals[i].Mask,
				Weight: uint32(bits.OnesCount32(evals[k].Mask | evals[i].Mask)),
				Cost:   evals[k].Cost + evals[i].Cost + uint32(bits.OnesCount32(evals[k].Mask&evals[i].Mask)),
				Level:  1 + max(evals[k].Level, evals[i].Level),
				Fan0:   k,
				Fan1:   i,
			}
			evals[nArgsNew] = eval
			if ivyMultiEvalBetter(eval, evals[bestIndex]) {
				bestIndex = nArgsNew
			}
			args = append(args, temp)
			nArgsNew++
			if nArgsNew == 15 {
				goto candidatesDone
			}
		}
	}

candidatesDone:
	if nArgsNew == int(nArgs) {
		Ivy_MultiSort(args, nArgs)
		return Ivy_MultiBalance_rec(p, args, nArgs, typ)
	}
	if nArgsNew == int(nArgs)+1 {
		if bestIndex != int(nArgs) {
			panic("ivy: unexpected best candidate index")
		}
		k := 0
		for i := 0; i < int(nArgs); i++ {
			if i != evals[bestIndex].Fan0 && i != evals[bestIndex].Fan1 {
				args[k] = args[i]
				k++
			}
		}
		args[k] = args[int(nArgs)]
		k++
		Ivy_MultiSort(args, int32(k))
		return Ivy_MultiBalance_rec(p, args, int32(k), typ)
	}

	all := uint32(1)<<uint(nArgs) - 1
	if evals[bestIndex].Mask == all {
		return Ivy_MultiBuild_rec(p, evals, bestIndex, args, int32(nArgsNew), typ)
	}

	nEvals := nArgsNew
	for i := 1; i < nArgsNew; i++ {
		for k := 0; k < i; k++ {
			eval := Ivy_Eval_t{
				Mask:   evals[k].Mask | evals[i].Mask,
				Weight: uint32(bits.OnesCount32(evals[k].Mask | evals[i].Mask)),
				Cost:   evals[k].Cost + evals[i].Cost + uint32(bits.OnesCount32(evals[k].Mask&evals[i].Mask)),
				Level:  1 + max(evals[k].Level, evals[i].Level),
				Fan0:   k,
				Fan1:   i,
			}
			if nEvals >= len(evals) {
				panic("ivy: multi-input candidate table overflow")
			}
			evals[nEvals] = eval
			if ivyMultiEvalBetter(eval, evals[bestIndex]) {
				bestIndex = nEvals
			}
			nEvals++
		}
	}

	bestNode := Ivy_MultiBuild_rec(p, evals, bestIndex, args, int32(nArgsNew), typ)
	k := 0
	for i := 0; i < int(nArgs); i++ {
		if evals[bestIndex].Mask&(uint32(1)<<uint(i)) == 0 {
			args[k] = args[i]
			k++
		}
	}
	args[k] = bestNode
	k++
	Ivy_MultiSort(args, int32(k))
	return Ivy_MultiBalance_rec(p, args, int32(k), typ)
}

func ivyMultiEvalBetter(candidate, best Ivy_Eval_t) bool {
	return candidate.Weight > best.Weight ||
		candidate.Weight == best.Weight && candidate.Cost < best.Cost ||
		candidate.Weight == best.Weight && candidate.Cost == best.Cost && candidate.Level < best.Level
}

func Ivy_MultiBuild_rec(p *Ivy_Man_t, evals []Ivy_Eval_t, index int, args []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if index < int(nArgs) {
		return args[index]
	}
	if index >= len(evals) {
		panic("ivy: candidate index outside evaluation table")
	}
	eval := evals[index]
	return Ivy_Oper(p,
		Ivy_MultiBuild_rec(p, evals, eval.Fan0, args, nArgs, typ),
		Ivy_MultiBuild_rec(p, evals, eval.Fan1, args, nArgs, typ), typ)
}

func Ivy_MultiSort(args []*Ivy_Obj_t, nArgs int32) {
	if nArgs < 0 || int64(nArgs) > int64(len(args)) {
		panic("ivy: invalid multi-input sort length")
	}
	for i := 0; i < int(nArgs)-1; i++ {
		best := i
		for j := i + 1; j < int(nArgs); j++ {
			if Ivy_ObjLevel(args[j]) > Ivy_ObjLevel(args[best]) {
				best = j
			}
		}
		args[i], args[best] = args[best], args[i]
	}
}

func Ivy_MultiPushUniqueOrderByLevel(args []*Ivy_Obj_t, nArgs int32, node *Ivy_Obj_t) int32 {
	if nArgs < 0 || int64(nArgs) >= int64(len(args)) {
		panic("ivy: multi-input work slice is too small")
	}
	for i := int32(0); i < nArgs; i++ {
		if args[i] == node {
			return nArgs
		}
	}
	args[nArgs] = node
	nArgs++
	for i := nArgs - 1; i > 0; i-- {
		if Ivy_ObjLevel(args[i]) <= Ivy_ObjLevel(args[i-1]) {
			break
		}
		args[i], args[i-1] = args[i-1], args[i]
	}
	return nArgs
}

func Ivy_MultiBalance_rec(p *Ivy_Man_t, args []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	assert(nArgs > 0, "nArgs > 0")
	if nArgs == 1 {
		return args[0]
	}
	if nArgs == 2 {
		return Ivy_Oper(p, args[0], args[1], typ)
	}
	node := Ivy_Oper(p, args[nArgs-1], args[nArgs-2], typ)
	nArgs = Ivy_MultiPushUniqueOrderByLevel(args, nArgs-2, node)
	return Ivy_MultiBalance_rec(p, args, nArgs, typ)
}

// Ivy_MultiEval retains ABC's diagnostic routine for inspecting reusable
// operations already present in the structural hash.
func Ivy_MultiEval(p *Ivy_Man_t, args []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if p == nil || nArgs < 0 || int64(nArgs) > int64(len(args)) {
		panic("ivy: invalid multi-input diagnostic arguments")
	}
	old := nArgs
	for i := int32(0); i < nArgs; i++ {
		fmt.Printf("%d[%d] ", i, Ivy_ObjLevel(args[i]))
	}
	for i := int32(1); i < nArgs; i++ {
		for k := int32(0); k < i; k++ {
			ghost, ok := ivyMultiGhost(p, args[k], args[i], typ)
			if !ok {
				continue
			}
			if temp := Ivy_TableLookup(p, ghost); temp != nil {
				fmt.Printf("%d[%d]=(%d,%d) ", nArgs, Ivy_ObjLevel(temp), k, i)
				args = append(args, temp)
				nArgs++
			}
		}
	}
	fmt.Printf("     ((%d/%d))    ", old, nArgs-old)
	return nil
}

func Ivy_Multi1(p *Ivy_Man_t, args []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if typ != IVY_AND && typ != IVY_EXOR || nArgs <= 0 || int64(nArgs) > int64(len(args)) {
		panic("ivy: invalid multi-input operation")
	}
	for _, arg := range args[:nArgs] {
		ivyCheckOperands(p, arg)
	}
	if nArgs == 1 {
		return args[0]
	}
	refs := make([]*Ivy_Obj_t, 0, nArgs)
	for _, arg := range args[:nArgs] {
		if Ivy_ObjRefs(Ivy_Regular(arg)) > 0 {
			refs = append(refs, arg)
		}
	}
	for i := 0; i < len(refs); i++ {
		for k := i + 1; k < len(refs); k++ {
			ghost, ok := ivyMultiGhost(p, refs[i], refs[k], typ)
			if !ok || Ivy_TableLookup(p, ghost) == nil {
				continue
			}
			combined := make([]*Ivy_Obj_t, 0, nArgs-1)
			for _, arg := range args[:nArgs] {
				if arg != refs[i] && arg != refs[k] {
					combined = append(combined, arg)
				}
			}
			combined = append(combined, Ivy_TableLookup(p, ghost))
			assert(int32(len(combined)) == nArgs-1, "len(combined) == nArgs-1")
			return Ivy_Multi1(p, combined, int32(len(combined)), typ)
		}
	}
	return Ivy_Multi_rec(p, args, nArgs, typ)
}

func Ivy_Multi2(p *Ivy_Man_t, args []*Ivy_Obj_t, nArgs int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if typ != IVY_AND && typ != IVY_EXOR {
		panic("ivy: unsupported multi-input operation")
	}
	return Ivy_Multi_rec(p, args, nArgs, typ)
}
