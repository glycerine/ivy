package pdr

import (
	"slices"

	"github.com/glycerine/abcgo/misc/vec"
)

// Pdr_Set_t holds sorted flop literals followed by the associated PI literals.
// Sign is the original 63-bit membership filter; nRefs tracks the PDR owner's
// references independently of Go's garbage collector.
type Pdr_Set_t struct {
	Sign                 uint64
	nRefs, nTotal, nLits int32
	Lits                 []int32
}

func Pdr_SetAlloc(nSize int32) *Pdr_Set_t {
	if nSize < 0 || nSize >= 1<<30 {
		panic("pdr: invalid cube size")
	}
	return &Pdr_Set_t{Lits: make([]int32, nSize)}
}

func Pdr_SetCreate(vLits, vPiLits *vec.Vec_Int_t) *Pdr_Set_t {
	flops, inputs := vec.Vec_IntArray(vLits), vec.Vec_IntArray(vPiLits)
	if len(flops)+len(inputs) >= 1<<30 {
		panic("pdr: cube too large")
	}
	p := Pdr_SetAlloc(int32(len(flops) + len(inputs)))
	p.nRefs, p.nLits, p.nTotal = 1, int32(len(flops)), int32(len(p.Lits))
	copy(p.Lits, flops)
	copy(p.Lits[len(flops):], inputs)
	slices.Sort(p.Lits[:p.nLits])
	for _, lit := range p.Lits[:p.nLits] {
		if lit < 0 {
			panic("pdr: negative cube literal")
		}
		p.Sign |= uint64(1) << uint(lit%63)
	}
	return p
}

func Pdr_SetCreateFrom(set *Pdr_Set_t, remove int32) *Pdr_Set_t {
	if remove < 0 || remove >= set.nLits {
		panic("pdr: invalid removed literal index")
	}
	lits := make([]int32, 0, set.nLits-1)
	lits = append(lits, set.Lits[:remove]...)
	lits = append(lits, set.Lits[remove+1:set.nLits]...)
	return pdrCreateSubset(set, lits)
}

func Pdr_SetCreateSubset(set *Pdr_Set_t, lits []int32, nLits int32) *Pdr_Set_t {
	if nLits < 0 || nLits > set.nLits || int(nLits) > len(lits) {
		panic("pdr: invalid subset size")
	}
	return pdrCreateSubset(set, lits[:nLits])
}

func pdrCreateSubset(set *Pdr_Set_t, lits []int32) *Pdr_Set_t {
	flops := vec.Vec_IntAllocArrayCopy(lits, len(lits))
	inputs := set.Lits[set.nLits:set.nTotal]
	return Pdr_SetCreate(flops, vec.Vec_IntAllocArrayCopy(inputs, len(inputs)))
}

func Pdr_SetDup(set *Pdr_Set_t) *Pdr_Set_t {
	copySet := *set
	copySet.Lits = slices.Clone(set.Lits)
	copySet.nRefs = 1
	return &copySet
}

func Pdr_SetRef(set *Pdr_Set_t) *Pdr_Set_t { set.nRefs++; return set }
func Pdr_SetDeref(set *Pdr_Set_t) {
	// ABC also uses -1 to mark noninductive invariant cubes. Its dereference
	// decrements that sentinel rather than treating it as an assertion failure.
	set.nRefs--
}

// Pdr_SetContains reports whether old contains every flop literal in subset.
func Pdr_SetContains(old, subset *Pdr_Set_t) int32 {
	if old.nLits < subset.nLits || old.Sign&subset.Sign != subset.Sign {
		return 0
	}
	return pdrContains(old, subset, false)
}

// Pdr_SetContainsSimple also accepts -1 placeholders in the subset.
func Pdr_SetContainsSimple(old, subset *Pdr_Set_t) int32 {
	return pdrContains(old, subset, true)
}

func pdrContains(old, subset *Pdr_Set_t, skipRemoved bool) int32 {
	i, j := int(old.nLits)-1, int(subset.nLits)-1
	for j >= 0 {
		if skipRemoved && subset.Lits[j] == -1 {
			j--
			continue
		}
		if i < 0 {
			return 0
		}
		switch {
		case old.Lits[i] == subset.Lits[j]:
			i--
			j--
		case old.Lits[i] > subset.Lits[j]:
			i--
		default:
			return 0
		}
	}
	return 1
}

func Pdr_SetIsInit(cube *Pdr_Set_t, remove int32) int32 {
	for i, lit := range cube.Lits[:cube.nLits] {
		if lit < 0 {
			panic("pdr: removed literal in cube")
		}
		if int32(i) != remove && lit&1 == 0 {
			return 0
		}
	}
	return 1
}

func Pdr_SetCompare(left, right **Pdr_Set_t) int32 {
	a, b := *left, *right
	for i := 0; i < int(a.nLits) && i < int(b.nLits); i++ {
		if a.Lits[i] > b.Lits[i] {
			return -1
		}
		if a.Lits[i] < b.Lits[i] {
			return 1
		}
	}
	if a.nLits < b.nLits {
		return -1
	}
	if a.nLits > b.nLits {
		return 1
	}
	return 0
}
