package gia

import "math/bits"

const (
	LF_LEAF_MAX  = 13
	LF_CUT_MAX   = 32
	LF_LOG_PAGE  = 12
	LF_NO_LEAF   = 255
	LF_CUT_WORDS = 4 + LF_LEAF_MAX/2
	LF_TT_WORDS  = 1 << (LF_LEAF_MAX - 6)
	LF_EPSILON   = 0.005
)

type Lf_Cut_t struct {
	Sign                        word
	Delay                       int32
	Flow                        float32
	iFunc                       int32
	Cost, fLate, fMux7, nLeaves uint32
	pLeaves                     [LF_LEAF_MAX]int32
}
type Lf_Plc_t struct{ fUsed, Handle uint32 }
type Lf_Bst_t struct {
	Delay [3]int32
	Flow  [3]float32
	Cut   [2]Lf_Plc_t
}

func Lf_BestDiffCuts(p *Lf_Bst_t) int32 {
	if p.Cut[0].Handle != p.Cut[1].Handle {
		return 1
	}
	return 0
}
func Lf_BestIsMapped(p *Lf_Bst_t) int32 { return int32(p.Cut[0].fUsed ^ p.Cut[1].fUsed) }
func Lf_BestIndex(p *Lf_Bst_t) int32    { return int32(p.Cut[1].fUsed) }
func Lf_BestCutIndex(p *Lf_Bst_t) int32 {
	if p.Cut[0].fUsed != 0 {
		return 0
	}
	if p.Cut[1].fUsed != 0 {
		return 1
	}
	return 2
}
func Lf_CutCopy(p, q *Lf_Cut_t, n int32) { *p = *q }
func Lf_CutCreateUnit(p *Lf_Cut_t, id int32) int32 {
	p.fLate = 0
	p.fMux7 = 0
	p.iFunc = 2
	p.nLeaves = 1
	p.pLeaves[0] = id
	p.Sign = word(1) << uint(id&63)
	return 1
}
func Lf_CutGetSign(c *Lf_Cut_t) word {
	var s word
	for _, id := range c.pLeaves[:c.nLeaves] {
		s |= word(1) << uint(id&63)
	}
	return s
}
func Lf_CutCountBits(x word) int32 { return int32(bits.OnesCount64(x)) }
func Lf_CutEqual(a, b *Lf_Cut_t) bool {
	if a.iFunc != b.iFunc || a.nLeaves != b.nLeaves {
		return false
	}
	for i := uint32(0); i < a.nLeaves; i++ {
		if a.pLeaves[i] != b.pLeaves[i] {
			return false
		}
	}
	return true
}
func Lf_CutCheck(base, cut *Lf_Cut_t) bool {
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		found := false
		for _, x := range base.pLeaves[:base.nLeaves] {
			if x == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func Lf_SetCutIsContainedOrder(base, cut *Lf_Cut_t) bool {
	if base.nLeaves < cut.nLeaves {
		return false
	}
	k := uint32(0)
	for _, v := range base.pLeaves[:base.nLeaves] {
		if k == cut.nLeaves {
			return true
		}
		if v > cut.pLeaves[k] {
			return false
		}
		if v == cut.pLeaves[k] {
			k++
		}
	}
	return k == cut.nLeaves
}
func lfCutMerge(cuts []*Lf_Cut_t, out *Lf_Cut_t, limit int32) int32 {
	indices := make([]uint32, len(cuts))
	n := int32(0)
	var sign word
	for _, c := range cuts {
		sign |= c.Sign
	}
	for {
		var v int32
		found := false
		for i, c := range cuts {
			if indices[i] < c.nLeaves && (!found || c.pLeaves[indices[i]] < v) {
				v = c.pLeaves[indices[i]]
				found = true
			}
		}
		if !found {
			break
		}
		if n == limit {
			return 0
		}
		out.pLeaves[n] = v
		n++
		for i, c := range cuts {
			if indices[i] < c.nLeaves && c.pLeaves[indices[i]] == v {
				indices[i]++
			}
		}
	}
	out.nLeaves = uint32(n)
	out.iFunc = -1
	out.fMux7 = 0
	out.Sign = sign
	return 1
}
func Lf_CutMergeOrder(a, b, out *Lf_Cut_t, limit int32) int32 {
	old := out.fMux7
	result := lfCutMerge([]*Lf_Cut_t{a, b}, out, limit)
	if result != 0 && int32(a.nLeaves) == limit && int32(b.nLeaves) == limit {
		out.fMux7 = old
	}
	return result
}
func Lf_CutMergeOrder2(a, b, out *Lf_Cut_t, limit int32) int32 {
	return lfCutMerge([]*Lf_Cut_t{a, b}, out, limit)
}
func Lf_CutMergeOrderMux(a, b, c, out *Lf_Cut_t, limit int32) int32 {
	return lfCutMerge([]*Lf_Cut_t{a, b, c}, out, limit)
}
func Lf_CutCompareDelay(a, b *Lf_Cut_t) int32 {
	if a.Delay < b.Delay {
		return -1
	}
	if a.Delay > b.Delay {
		return 1
	}
	if a.nLeaves < b.nLeaves {
		return -1
	}
	if a.nLeaves > b.nLeaves {
		return 1
	}
	if a.Flow < b.Flow-LF_EPSILON {
		return -1
	}
	if a.Flow > b.Flow+LF_EPSILON {
		return 1
	}
	return 0
}
func Lf_CutCompareArea(a, b *Lf_Cut_t) int32 {
	if a.fLate < b.fLate {
		return -1
	}
	if a.fLate > b.fLate {
		return 1
	}
	if a.Flow < b.Flow-LF_EPSILON {
		return -1
	}
	if a.Flow > b.Flow+LF_EPSILON {
		return 1
	}
	if a.Delay < b.Delay {
		return -1
	}
	if a.Delay > b.Delay {
		return 1
	}
	if a.nLeaves < b.nLeaves {
		return -1
	}
	if a.nLeaves > b.nLeaves {
		return 1
	}
	return 0
}

func Lf_SetCheckArray(cuts []*Lf_Cut_t, n int32) int32 {
	if n <= 0 {
		panic("LF empty cut set")
	}
	for i := int32(0); i < n; i++ {
		c := cuts[i]
		if c.fMux7 != 0 || c.nLeaves >= LF_LEAF_MAX || c.Sign != Lf_CutGetSign(c) {
			panic("LF invalid cut")
		}
		for a := uint32(0); a < c.nLeaves; a++ {
			for b := a + 1; b < c.nLeaves; b++ {
				if c.pLeaves[a] >= c.pLeaves[b] {
					panic("LF unsorted leaves")
				}
			}
		}
		for k := int32(0); k < n; k++ {
			if i != k && Lf_CutCheck(c, cuts[k]) {
				panic("LF dominated cut")
			}
		}
	}
	return 1
}
func Lf_SetLastCutIsContained(cuts []*Lf_Cut_t, n int32) int32 {
	c := cuts[n]
	for i := int32(0); i < n; i++ {
		other := cuts[i]
		if other.nLeaves <= c.nLeaves && other.Sign&c.Sign == other.Sign && Lf_SetCutIsContainedOrder(c, other) {
			return 1
		}
	}
	return 0
}
func Lf_SetLastCutContainsArea(cuts []*Lf_Cut_t, n int32) int32 {
	c := cuts[n]
	changed := false
	for i := int32(1); i < n; i++ {
		other := cuts[i]
		if c.nLeaves < other.nLeaves && c.Sign&other.Sign == c.Sign && Lf_SetCutIsContainedOrder(other, c) {
			other.nLeaves = LF_NO_LEAF
			changed = true
		}
	}
	if !changed {
		return n
	}
	k := int32(1)
	for i := int32(1); i <= n; i++ {
		if cuts[i].nLeaves == LF_NO_LEAF {
			continue
		}
		if k < i {
			cuts[k], cuts[i] = cuts[i], cuts[k]
		}
		k++
	}
	return k - 1
}
func Lf_SetSortByArea(cuts []*Lf_Cut_t, n int32) {
	for i := n; i > 1; i-- {
		if Lf_CutCompareArea(cuts[i-1], cuts[i]) < 0 {
			return
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
}
func Lf_SetAddCut(cuts []*Lf_Cut_t, n, limit int32) int32 {
	if n == 0 {
		return 1
	}
	n = Lf_SetLastCutContainsArea(cuts, n)
	if n < 1 {
		panic("LF lost delay cut")
	}
	if Lf_CutCompareDelay(cuts[0], cuts[n]) == 1 {
		cuts[0], cuts[n] = cuts[n], cuts[0]
		if cuts[0].nLeaves < cuts[n].nLeaves && cuts[0].Sign&cuts[n].Sign == cuts[0].Sign && Lf_SetCutIsContainedOrder(cuts[n], cuts[0]) {
			return n
		}
	}
	Lf_SetSortByArea(cuts, n)
	return min(n+1, limit-1)
}
func Lf_SetSortBySize(cuts []*Lf_Cut_t, n int32) {
	for i := int32(1); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if cuts[j].nLeaves > cuts[best].nLeaves {
				best = j
			}
		}
		cuts[i], cuts[best] = cuts[best], cuts[i]
	}
}
