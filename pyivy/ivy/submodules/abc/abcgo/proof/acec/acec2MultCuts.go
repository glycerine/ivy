package acec

// Cut set operations from src/proof/acec/acec2Mult.c.
func Sdb_CutGetSign(cut *Sdb_Cut_t) uint64 {
	var sign uint64
	for _, leaf := range cut.pLeaves[:cut.nLeaves] {
		sign |= uint64(1) << uint32(leaf&63)
	}
	return sign
}
func Sdb_CutCheck(base, cut *Sdb_Cut_t) int32 {
	for _, c := range cut.pLeaves[:cut.nLeaves] {
		found := false
		for _, b := range base.pLeaves[:base.nLeaves] {
			if c == b {
				found = true
				break
			}
		}
		if !found {
			return 0
		}
	}
	return 1
}
func Sdb_CutSetCheckArray(cuts []*Sdb_Cut_t, n int32) int32 {
	if n <= 0 {
		panic("acec: empty cut set")
	}
	for i := int32(0); i < n; i++ {
		a := cuts[i]
		if a.nLeaves > SDB_MAX_CUTSIZE || a.Sign != Sdb_CutGetSign(a) {
			panic("acec: invalid cut")
		}
		for j := uint32(1); j < a.nLeaves; j++ {
			if a.pLeaves[j-1] >= a.pLeaves[j] {
				panic("acec: unsorted cut")
			}
		}
		for j := int32(0); j < n; j++ {
			if a != cuts[j] && Sdb_CutCheck(a, cuts[j]) != 0 {
				panic("acec: contained cut")
			}
		}
	}
	return 1
}
func Sdb_CutMergeOrder(a, b, out *Sdb_Cut_t, limit int32) int32 {
	// Both inputs are sorted sets. A failed merge leaves the cut metadata untouched.
	i, j, k := uint32(0), uint32(0), uint32(0)
	for i < a.nLeaves || j < b.nLeaves {
		if int32(k) == limit {
			return 0
		}
		switch {
		case i == a.nLeaves:
			out.pLeaves[k] = b.pLeaves[j]
			j++
		case j == b.nLeaves:
			out.pLeaves[k] = a.pLeaves[i]
			i++
		case a.pLeaves[i] < b.pLeaves[j]:
			out.pLeaves[k] = a.pLeaves[i]
			i++
		case a.pLeaves[i] > b.pLeaves[j]:
			out.pLeaves[k] = b.pLeaves[j]
			j++
		default:
			out.pLeaves[k] = a.pLeaves[i]
			i++
			j++
		}
		k++
	}
	out.nLeaves = k
	out.iFunc = -1
	out.Sign = a.Sign | b.Sign
	return 1
}
func Sdb_CutMergeOrder2(a, b, out *Sdb_Cut_t, limit int32) int32 {
	return Sdb_CutMergeOrder(a, b, out, limit)
}
func Sdb_CutSetCutIsContainedOrder(base, cut *Sdb_Cut_t) int32 {
	if base.nLeaves < cut.nLeaves {
		panic("acec: invalid containment order")
	}
	i, j := uint32(0), uint32(0)
	for i < base.nLeaves && j < cut.nLeaves {
		if base.pLeaves[i] > cut.pLeaves[j] {
			return 0
		}
		if base.pLeaves[i] == cut.pLeaves[j] {
			j++
		}
		i++
	}
	if j == cut.nLeaves {
		return 1
	}
	return 0
}
func Sdb_CutSetLastCutIsContained(cuts []*Sdb_Cut_t, n int32) int32 {
	last := cuts[n]
	for i := int32(0); i < n; i++ {
		c := cuts[i]
		if c.nLeaves <= last.nLeaves && c.Sign&last.Sign == c.Sign && Sdb_CutSetCutIsContainedOrder(last, c) != 0 {
			return 1
		}
	}
	return 0
}
func Sdb_CutCompare(a, b *Sdb_Cut_t) int32 {
	if a.nTreeLeaves < b.nTreeLeaves {
		return -1
	}
	if a.nTreeLeaves > b.nTreeLeaves {
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
func Sdb_CutSetLastCutContains(cuts []*Sdb_Cut_t, n int32) int32 {
	last := cuts[n]
	changed := false
	for i := int32(0); i < n; i++ {
		c := cuts[i]
		if last.nLeaves < c.nLeaves && last.Sign&c.Sign == last.Sign && Sdb_CutSetCutIsContainedOrder(c, last) != 0 {
			c.nLeaves = SDB_CUT_NO_LEAF
			changed = true
		}
	}
	if !changed {
		return n
	}
	k := int32(0)
	for i := int32(0); i <= n; i++ {
		if cuts[i].nLeaves == SDB_CUT_NO_LEAF {
			continue
		}
		if k < i {
			cuts[k], cuts[i] = cuts[i], cuts[k]
		}
		k++
	}
	return k - 1
}
func Sdb_CutSetSortByCost(cuts []*Sdb_Cut_t, n int32) {
	for i := n; i > 0; i-- {
		if Sdb_CutCompare(cuts[i-1], cuts[i]) < 0 {
			return
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
}
func Sdb_CutSetAddCut(cuts []*Sdb_Cut_t, n, limit int32) int32 {
	if n == 0 {
		return 1
	}
	n = Sdb_CutSetLastCutContains(cuts, n)
	Sdb_CutSetSortByCost(cuts, n)
	if n+1 < limit-1 {
		return n + 1
	}
	return limit - 1
}
